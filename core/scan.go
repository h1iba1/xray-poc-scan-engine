package core

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/net/html/charset"
	"golang.org/x/sync/semaphore"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
	"xray-poc-scan-engine/utils"
)

// poc执行结构体
type PocExecuteManager struct {
	reqChan              chan *PocRequest
	resultChan           chan *PocResult
	wg                   sync.WaitGroup
	semWg                sync.WaitGroup
	sem                  *semaphore.Weighted
	running              int32
	httpClient           *http.Client
	noRedirectHttpClient *http.Client //不跟随重定向的client
	maxBodySize          int
	userAgents           []string
	showResponseBody     bool
	dialTimeout          int
	readTimeout          int
	//httpConfig           HttpConfig
	jar                  *cookiejar.Jar
	jar2                 *cookiejar.Jar
}

// poc执行管理
func NewPocExecuteManager(httpConfig HttpConfig) *PocExecuteManager {
	var (
		httpClient  *http.Client
		httpClient2 *http.Client
		err         error
		maxBodySize int
	)

	if httpConfig.MaxRespBodySize < 1024*1024 {
		maxBodySize = 1024 * 1024
	} else {
		maxBodySize = httpConfig.MaxRespBodySize
	}

	jar, _ := cookiejar.New(nil)
	httpClient, err = utils.NewHttpClient(&utils.HttpOptions{
		DialTimeout:     httpConfig.DialTimeout,
		ReadTimeout:     httpConfig.ReadTimeout,
		//Proxy:           httpConfig.Proxy,
		NeedRedirect:    true,
		MaxConnsPerHost: httpConfig.MaxConnsPerHost,
		MaxQPS:          httpConfig.MaxQPS,
		Jar:             jar,
	})
	if err != nil {
		panic(err)
	}
	jar2, _ := cookiejar.New(nil)
	// 次httpClient不进行重定向
	httpClient2, err = utils.NewHttpClient(&utils.HttpOptions{
		DialTimeout:     httpConfig.DialTimeout,
		ReadTimeout:     httpConfig.ReadTimeout,
		//Proxy:           httpConfig.Proxy,
		NeedRedirect:    false,
		MaxConnsPerHost: httpConfig.MaxConnsPerHost,
		MaxQPS:          httpConfig.MaxQPS,
		Jar:             jar2,
	})
	if err != nil {
		panic(err)
	}

	p := &PocExecuteManager{
		reqChan:              make(chan *PocRequest, 20),
		resultChan:           make(chan *PocResult, 20),
		sem:                  semaphore.NewWeighted(10),
		running:              1,
		httpClient:           httpClient,
		noRedirectHttpClient: httpClient2,
		maxBodySize:          maxBodySize,
		userAgents:           httpConfig.Headers.UserAgent,
		dialTimeout:          httpConfig.DialTimeout,
		readTimeout:          httpConfig.ReadTimeout,
		jar:                  jar,
		jar2:                 jar2,
	}

	return p
}


func (p *PocExecuteManager) Start() <-chan *PocResult {
	p.wg.Add(1)
	go p.run()
	return p.resultChan
}

func (p *PocExecuteManager) Close() {
	if atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		close(p.reqChan)
		p.wg.Wait()
	}
}

func (p *PocExecuteManager) AddRequest(req *PocRequest) {
	if atomic.LoadInt32(&p.running) == 1 {
		p.reqChan <- req
	}
}

func (p *PocExecuteManager) run() {
	defer func() {
		p.wg.Done()
	}()

	for req := range p.reqChan {
		if err := p.sem.Acquire(context.TODO(), 1); err == nil {
			p.semWg.Add(1)
			go p.scan(req)
		}
	}
	p.semWg.Wait()
}


// 根据pocYaml 发送请求 获取response
func (p *PocExecuteManager) protocolSend(pocYaml PocYaml, client *http.Client, request *http.Request, host, body string, port int) (*http.Response, []byte, error) {
	if pocYaml.Transport == "" || pocYaml.Transport == "http" {
		response, err := client.Do(request)
		if err !=nil{
			Log.Errorf("http do error %v",err)
		}
		return response,nil,nil
	} else if pocYaml.Transport == "tcp" || pocYaml.Transport == "udp" {

		dialer := &net.Dialer{
			Timeout: time.Duration(p.dialTimeout) * time.Second,
		}
		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		var conn net.Conn
		var err error
		buff := make([]byte, 1024)
		conn, err = tls.DialWithDialer(dialer, pocYaml.Transport, fmt.Sprintf("%s:%d", host, port), conf)
		if err != nil {
			if err.Error() == "tls: first record does not look like a TLS handshake" {
				conn, err = net.DialTimeout(pocYaml.Transport, fmt.Sprintf("%s:%d", host, port), time.Duration(p.dialTimeout)*time.Second)
			}
			if err != nil {
				return nil, nil, err
			}
		}

		defer conn.Close()

		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(p.readTimeout) * time.Second))
		_ = conn.SetWriteDeadline(time.Now().Add(time.Duration(p.readTimeout) * time.Second))

		var data []byte
		data = []byte(body)
		Log.Debugf("send data: %v", data)
		_, _ = conn.Write(data)
		_, err = conn.Read(buff)
		Log.Debugf("recv data: %v", buff)

			if err != nil {
				return nil, nil, err
			} else {
				return nil, buff, nil
			}
		}

	return nil, nil, errors.New("protocol not supported")
}

func (p *PocExecuteManager) scan(req *PocRequest) {
	defer func() {
		p.semWg.Done()
		p.sem.Release(1)
	}()

	result, err := p.DoScan(req)
	if err != nil {
		return
	}

	if result != nil {
		p.resultChan <- result
	}
}

// 核心执行函数
func (p *PocExecuteManager) DoScan(req *PocRequest)(*PocResult,error)  {
//	 报错处理
	defer func() {
		if err:=recover();err!=nil{
			Log.Errorf("panic error %v %v",err,string(debug.Stack()))
		}
	}()

	var(
		matchResult  = map[string]string{}
		find         = false
		pocYaml *PocYaml
		// sets变量存储
		setTypes =make(map[string]ref.Val)
		client       *http.Client

		// rule expression
		expResult    ExpResult
		param        PocParam
		responseBody []byte
		responseUrl  string
	)

	// 获取pocYaml
	pocYaml = GetPocYaml(req.YmlName)
	if pocYaml==nil{
		return nil,fmt.Errorf("can not find yml poc %s",req.YmlName)
	}

//	set参数处理
	pocYaml.Sets= parseYamlSets(*pocYaml)
	// 解析target url
	parts, err := url.Parse(req.URL)
	if err != nil {
		Log.Errorf("url parse error %v", err)
	}
	//根据schame host port 组合url
	//baseURL := fmt.Sprintf("%s://%s", parts.Scheme, parts.Host)
	pathURL := fmt.Sprintf("%s://%s%s", parts.Scheme, parts.Host, utils.URIPath(parts.Path))
	host := parts.Hostname()

	defaultVars := map[string]interface{}{
		"request.url.scheme":   parts.Scheme,
		"request.url.domain":   parts.Hostname(),
		"request.url.host":     parts.Host,
		"request.url.port":     parts.Port(),
		"request.url.path":     parts.Path,
		"request.url.query":    parts.RawQuery,
		"request.url.fragment": parts.Fragment,
	}

	pocResult := &PocResult{
		StartTime: time.Now(),
		// 复制pocYaml 的Detail
		Detail:    pocYaml.Detail,

		// 目标基础信息
		Target: PocTarget{
			Host:  parts.Hostname(),
			Typ:   "web",
			URL:   req.URL,
			Port:  utils.GetHttpPort(parts),
			Param: make([]PocParam, 0, len(pocYaml.Rules)),
		},
		ExpResult: expResult,
		Expression: pocYaml.Expression,
		Vulnerable: false,
	}

	// 处理sets中的cel表达式 将值存储在setTypes中
	celParseYamlSets(pocYaml.Sets,setTypes,defaultVars)

	// 处理expression
	expResult = expressionSlice(pocYaml.Expression)
//Exit:
	for i,exp:=range expResult.ExpResultSlice {

		var (
			ruleName string
		)
		ruleName =strings.Replace(exp.RuleName,"()","",-1)
		ruleName =strings.Replace(ruleName," ","",-1)
		rule :=pocYaml.Rules[ruleName]

		// 替换sets值
		pocYaml.Rules[ruleName]= ruleReplaceSet(&rule,setTypes)
		// 替换自定义规则查找到的值
		if len(matchResult) > 0 {
			pocYaml.Rules[ruleName]= ruleReplaceMatch(&rule,matchResult)
		}

		var(
			body =rule.Request.Body
			path =rule.Request.Path
			request      *http.Request
		)

		// 替换函数 将rule中所有的设置参数替换


	//	targetURL处理
		var targetURL string
		targetURL = pathURL + path

	//	 非get请求或者body!=nil 。处理request
		if rule.Request.Method !=http.MethodGet && body!=""{
			request,err=http.NewRequest(rule.Request.Method,targetURL,strings.NewReader(body))
		}else{
			request,err=http.NewRequest(rule.Request.Method,targetURL,nil)
		}
		if err!=nil{
			Log.Errorf("http new request error %v",err)
			return nil,err
		}

	// userAgents处理
		if len(p.userAgents) > 0{
			request.Header.Set("User-Agent", p.userAgents[0])
		}

	//	PocVerify
		var verify PocVerify
		verify.Payload=body
		// poc协议处理
		if pocYaml.Transport=="" || pocYaml.Transport=="http"{
			requestBody ,_:=httputil.DumpRequest(request,true)
			verify.HttpRaw.Request=string(requestBody)
		}else{
			// 非http协议
			verify.HttpRaw.Request=pocYaml.Transport+ "\r\n" + body
		}

	//	检查是否重定向
		if rule.Request.FollowRedirects{
			client =p.httpClient
		}else{
			client = p.noRedirectHttpClient
		}

	//	开始进行请求
		response, socketBody, err :=p.protocolSend(*pocYaml,client,request,host,body,req.Port)
		if err!=nil{
			Log.Errorf("protocol send error %v",err)
			return nil,err
		}

		if pocYaml.Transport=="" || pocYaml.Transport=="http"{
			if response != nil {
				// DumpResponse 类似于 DumpRequest 但转储响应。
				respBody, _ := httputil.DumpResponse(response, false)

				verify.HttpRaw.Response = string(respBody)
				responseBody, err = GetHTTPUtf8Body(response, p.maxBodySize)
				if err != nil {
					Log.Errorf("get http body error %v", err)
					return nil, err
				}
			}
		}else{
			verify.HttpRaw.Response = string(socketBody)
			responseBody = socketBody
		}

		// 将请求结果赋值给PocVerify
		verify.HttpRaw.Response += string(responseBody)

		param.Verify = []PocVerify{verify}

		Log.Debugf("request:%s ", verify.HttpRaw.Request)
		if p.showResponseBody {
			Log.Debugf("response:%s", verify.HttpRaw.Response)
		}

	//	处理search
		if rule.Output.Search!=""{

			if rule.Output.SearchReg==nil{
				// Compile 解析正则表达式，如果成功，则返回可用于匹配文本的 Regexp 对象。
				reg,err :=regexp.Compile(strings.Trim(strings.Trim(rule.Output.Search, "\r"), "\n"))
				if err != nil {
					Log.Errorf("search can not compile %v", err)
					return nil, err
				}
				rule.Output.SearchReg = reg
			}
			// SubexpNames 返回此 Regexp 中带括号的子表达式的名称。第一个子表达式的名称是 names[1]，因此如果 m 是匹配切片，则 m[i] 的名称是 SubexpNames()[i]。由于无法命名整个 Regexp，因此 names[0] 始终为空字符串。不应修改切片。
			groupNames :=rule.Output.SearchReg.SubexpNames()
			groupNamesLen :=len(groupNames)
			if groupNamesLen>0{
				matches := rule.Output.SearchReg.FindStringSubmatch(verify.HttpRaw.Response)
				Log.Debugf("matches %v", matches)
				for k, v := range groupNames {
					if v != "" && k < len(matches) {
						matchValue := matches[k]
						matchValue = strings.Trim(matchValue, " ")
						matchValue = strings.ReplaceAll(matchValue, `\\`, `\`)
						matchValue = strings.ReplaceAll(matchValue, `\`, `\\`) // 处理window路径问题
						matchResult[v] = matchValue
						// match 的结果作为自定义参数会传入到expression中
						setTypes[v] = types.String(matchValue)
					} else {
						if v != "" {
							setTypes[v] = types.String("")
						}
					}
				}
				Log.Debugf("matchResult %v %v", matchResult, i)
			}
		}

		// cel表达式
		celEnv, err := NewCelEnv(setTypes)
		if err != nil {
			Log.Errorf("create cel env error %v", err)
			return nil, err
		}
		pAst, iss := celEnv.Parse(rule.Expression)
		if iss != nil && iss.Err() != nil {
			err = iss.Err()
			Log.Errorf("rule expression parse error %v", err)
			return nil, err
		}
		cAst, iss := celEnv.Check(pAst)
		if iss != nil && iss.Err() != nil {
			err = iss.Err()
			Log.Errorf("rule expression check error %v", err)
			return nil, err
		}

		httpHeaders := make(map[string]string)
		statusCode :=0
		contentType := ""
		if response!=nil{
			for k,v :=range response.Header{
				httpHeaders[strings.ToLower(k)]=strings.Join(v,",")
			}

			responseUrl=response.Request.URL.String()
			statusCode =response.StatusCode
			contentType =response.Header.Get("Content-Type")
		}

		vars := map[string]interface{}{
			"response.status":       statusCode,
			"response.body":         responseBody,
			"response.content_type": contentType,
			"response.headers":      httpHeaders,
			"response.url":          responseUrl,
		}

		// 将poc的set参数添加到vars
		for k,v :=range setTypes{
			vars[k] = v.Value()
		}

		prg,err :=celEnv.Program(cAst, NewCelFunctions(""))
		if err!=nil{
			Log.Errorf("rule expression program error %v",err)
		}
		res,_,err :=prg.Eval(vars)
		if err != nil {
			Log.Errorf("rule expression eval error %v", err)
			return nil, err
		}

		Log.Debugf("res %v %s", res, rule.Expression)

		// expression结果判断
		if res.Value().(bool) {
			Log.Debugf("poc:%s rule url:%s execute success", pocYaml.Name, targetURL)
			find = true
			expResult.ExpResultSlice[i].RuleName=ruleName
			expResult.ExpResultSlice[i].RuleResult=find

			pocResult.Target.URL = targetURL
			pocResult.Target.Param = append(pocResult.Target.Param, param)
		} else {
			find = false
			expResult.ExpResultSlice[i].RuleName=ruleName
			expResult.ExpResultSlice[i].RuleResult=find
			Log.Debugf("poc:%s rule url:%s expression execute failed", pocYaml.Name, targetURL)
		}
	}

	vars := make(map[string]interface{})
	for k, v := range setTypes {
		vars[k] = v.Value()
	}

	pocResult.Outputs = parseYamlSetsWithOutput(pocYaml.Outputs, setTypes, vars)
	pocResult.ExpResult=expResult
	pocResult.PocInfo= PocInfo{
		PocName:pocYaml.Name,
	}

	// rule执行结果 综合判断
	ruleExecuteResultJudge(pocResult)
	return pocResult, nil

}

// 解析pocYaml 中的set参数
func parseYamlSets(pocYaml PocYaml) ProbesKV {
	setVars :=pocYaml.Sets
	for k,v :=range pocYaml.Set{
		setVars=append(setVars, KV{
			Key: k,
			Value: v,
		})
	}
	return setVars
}

// 将rule中的所有set替换
func ruleReplaceSet(rule *Rules,setTypes map[string]ref.Val) Rules {
	for k,v := range setTypes {
		// path
		if strings.Contains(rule.Request.Path,k){
			rule.Request.Path=strings.Replace(rule.Request.Path,"{{"+k+"}}",fmt.Sprintf("%v",v.Value()),-1)
		}

		//// body
		if strings.Contains(rule.Request.Body,k){
			rule.Request.Body=strings.Replace(rule.Request.Body,"{{"+k+"}}",fmt.Sprintf("%v",v.Value()),-1)
		}
		//// heades
		for k1,header :=range rule.Request.Headers{
			if strings.Contains(header,k){
				rule.Request.Headers[k1]=strings.Replace(rule.Request.Headers[k1],"{{"+k+"}}",fmt.Sprintf("%v",v.Value()),-1)
			}
		}
	}
	return *rule
}

// 替换自定义规则提取的值
func ruleReplaceMatch(rule *Rules,matchResult map[string]string) Rules {
	for k,v :=range matchResult{
		// path
		if strings.Contains(rule.Request.Path,k){
			rule.Request.Path=strings.Replace(rule.Request.Path,"{{"+k+"}}",v,-1)
		}

		// body
		if strings.Contains(rule.Request.Body,k){
			rule.Request.Body=strings.Replace(rule.Request.Body,"{{"+k+"}}",v,-1)
		}
		// heades
		for k1,header :=range rule.Request.Headers{
			if strings.Contains(header,k){
				rule.Request.Headers[k1]=strings.Replace(rule.Request.Headers[k1],"{{"+k+"}}",v,-1)
			}
		}
	}

	return *rule
}


// 解析pocYaml 中的set参数
func celParseYamlSets(sets ProbesKV, setTypes map[string]ref.Val, defaultVars map[string]interface{}) error {
	if len(sets) == 0 {
		return nil
	}
	for _, v := range sets {
		// 创建一个cel表达式
		defaultCelEnv, err := NewCelEnv(setTypes)
		if err != nil {
			Log.Errorf("%v",err)
			return err
		}
		// 将v.Value解析为ast
		pAst, iss := defaultCelEnv.Parse(v.Value)
		if iss != nil && iss.Err() != nil {
			return iss.Err()
		}
		//Check 对输入 Ast 执行类型检查，并产生一个经过检查的 Ast 和或一组问题。
		ast, iss := defaultCelEnv.Check(pAst)
		if iss != nil && iss.Err() != nil {
			return iss.Err()
		}
		// 在环境 (Env) 中生成 Ast 的可评估实例。
		prg, err := defaultCelEnv.Program(ast, NewCelFunctions(""))
		if err != nil {
			return err
		}
		// cel表达式解析参数
		rv, _, err := prg.Eval(defaultVars)
		if err != nil {
			return err
		}
		// 参数键值对
		setTypes[v.Key] = rv
		// 请求参数对应到defaultVars默认参数
		defaultVars[v.Key] = rv.Value()
	}
	// 输出参数设置结果
	Log.Debugf("setTypes values %v", setTypes)
	return nil
}

func expressionSlice(expression string) ExpResult {
	// rule执行结果
	var expResult ExpResult
	var expMap RuleResultJudge
	if strings.Contains(expression,"&&"){
		expSlice :=strings.Split(expression,"&&")

		for _,v:=range expSlice{
			expMap.RuleName=v
			expMap.RuleResult=false
			expResult.ExpResultSlice =append(expResult.ExpResultSlice,expMap)
		}
		return expResult
	}

	if strings.Contains(expression,"||"){
		expSlice :=strings.Split(expression,"||")

		for _,v:=range expSlice{
			expMap.RuleName=v
			expMap.RuleResult=false
			expResult.ExpResultSlice =append(expResult.ExpResultSlice,expMap)
		}
		return expResult
	}

	// TODO: 同时存在&& ||

	expMap.RuleName=expression
	expMap.RuleResult=false
	expResult.ExpResultSlice =append(expResult.ExpResultSlice,expMap)
	return expResult
}

// GetHTTPUtf8Body 将response转为utf-8
func GetHTTPUtf8Body(response *http.Response, maxBodySize int) ([]byte, error) {
	defer response.Body.Close()
	body, err := GetHTTPOriginalBody(response, maxBodySize)
	if err != nil {
		Log.Debugf("%v",err)
		return nil, err
	}
	body, _ = ForceHtmlUtf8(body, string(response.Header.Get("Content-Type")))
	return body, nil
}

// GetHTTPOriginalBody 获取 HTTP 原始正文
func GetHTTPOriginalBody(response *http.Response, maxBodySize int) ([]byte, error) {
	var err error
	// 返回一个指定大小的response.Body
	bodyReader := io.LimitReader(response.Body, int64(maxBodySize))
	// 获取响应头Content-Encoding的值 并转为小写
	contentEncoding := strings.ToLower(response.Header.Get("Content-Encoding"))
	// 如果response 未压缩
	if !response.Uncompressed {
		// Content-Encoding==gzip
		if contentEncoding == "gzip" {
			bodyReader, err = gzip.NewReader(bodyReader)
			if err != nil {
				return nil, err
			}
		} else if contentEncoding == "deflate" {
			bodyReader, err = zlib.NewReader(bodyReader)
			if err != nil {
				return nil, err
			}
		}
	}

	// ReadAll 被定义为从 src 读取直到 EOF。并返回[]byte
	body, _ := ioutil.ReadAll(bodyReader)
	return body, nil
}

func ForceHtmlUtf8(body []byte, contentType string) ([]byte, string) {
	htmlCharset := getCharSet(contentType)
	if htmlCharset == "" {
		htmlCharset = detectHtmlCharset(body)
		if htmlCharset == "" {
			_, htmlCharset, _ = charset.DetermineEncoding(body, contentType)
		}
	}

	return ForceUtf8(body, htmlCharset)
}

// 获取contentType charset=值
func getCharSet(contentType string) string {
	defer func() {
		if err := recover(); err != nil {
			Log.Errorf("getCharSet error %v %v", err, contentType)
		}
	}()
	content := strings.ToLower(contentType)

	pos := strings.Index(content, "charset=")
	if pos > 0 {
		begin := pos + 8
		if len(contentType) >= begin {
			return strings.TrimSpace(contentType[begin:])
		}
	}
	return ""
}

var charsetPattern = regexp.MustCompile(`(?i)<meta[^>]+charset\s*=\s*["]{0,1}([a-z0-9-]*)`)
// 检测html字符集
func detectHtmlCharset(body []byte) string {
	if len(body) > 1024 {
		body = body[:1024]
	}
	match := charsetPattern.FindSubmatch(body)
	if match == nil {
		return ""
	}
	return string(match[1])
}

// ForceUtf8 转为utf-8字符集
func ForceUtf8(body []byte, charsetName string) ([]byte, string) {
	if strings.ToLower(charsetName) == "utf-8" && utf8.Valid(body) {
		return body, charsetName
	}

	reader := transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		Log.Errorf("%v",e)
		return body, charsetName
	}
	return d, charsetName
}

func parseYamlSetsWithOutput(sets ProbesKV, setTypes map[string]ref.Val, defaultVars map[string]interface{}) map[string]string {
	result := make(map[string]string)
	if len(sets) == 0 {
		return result
	}
	for _, v := range sets {
		defaultCelEnv, err := NewCelEnv(setTypes)
		if err != nil {
			Log.Errorf("%v",err)
			return result
		}
		pAst, iss := defaultCelEnv.Parse(v.Value)
		if iss != nil && iss.Err() != nil {
			return result
		}
		ast, iss := defaultCelEnv.Check(pAst)
		if iss != nil && iss.Err() != nil {
			return result
		}
		prg, err := defaultCelEnv.Program(ast, NewCelFunctions(""))
		if err != nil {
			return result
		}

		rv, _, err := prg.Eval(defaultVars)
		if err != nil {
			return result
		}
		setTypes[v.Key] = rv
		defaultVars[v.Key] = rv.Value()
		result[v.Key] = fmt.Sprintf("%v", rv.Value())
	}
	return result
}

// rule执行结果综合判断
func ruleExecuteResultJudge(pocResult *PocResult) {
	//	只有单独一个rule
	if len(pocResult.ExpResult.ExpResultSlice)==1{
		if pocResult.ExpResult.ExpResultSlice[0].RuleResult==true{
			pocResult.Vulnerable =true
		}else if pocResult.ExpResult.ExpResultSlice[0].RuleResult==false{
			pocResult.Vulnerable =false
		}
	}

	// && 有一个为flase 则整体都为flase
	if strings.Contains(pocResult.Expression,"&&"){
		pocResult.Vulnerable =true
		for _,v :=range pocResult.ExpResult.ExpResultSlice {
			if v.RuleResult==false{
				pocResult.Vulnerable =false
			}
		}
	}

	// || 有一个为true则为true
	if strings.Contains(pocResult.Expression,"||"){
		for _,v :=range pocResult.ExpResult.ExpResultSlice {
			if v.RuleResult==true{
				pocResult.Vulnerable =true
			}
		}
	}

//	TODO || && 一起时的逻辑规则
}