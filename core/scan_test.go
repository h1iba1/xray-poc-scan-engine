package core

import (
	"fmt"
	"github.com/google/cel-go/common/types/ref"
	"net/http"
	"net/url"
	"testing"
)

func init() {
	err := LoadPackPoc()
	if err!=nil{}
	LoadAllYamlPoc()
}


func TestDoScan(t *testing.T) {

	req := &PocRequest{
		URL: "https://www.h11ba1.com",
		YmlName: "poc-yaml-yapi-rce",
		Headers: http.Header{},
		Method: "get",
		PostData:"",
		Port:80,
	}
	userAgent :=[]string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169"}
	httpConfig := HttpConfig{
		DialTimeout:     5,
		ReadTimeout:     30,
		FailRetries:     1,
		MaxQPS:          500,
		MaxRedirect:     5,
		MaxConnsPerHost: 50,
		MaxRespBodySize: 8388608,
		Headers:         HeaderConfig{UserAgent: userAgent},
	}

	pocExecuteManager := NewPocExecuteManager(httpConfig)

	var pocResult *PocResult

	pocResult,_ =pocExecuteManager.DoScan(req)
	//if err != nil {
	//	log.Errorf("poc执行错误 %v",err)
	//}

	//for i,exp :=range pocResult.ExpResult.expResultSlice{
	//	fmt.Printf("k %v\n", i)
	//	fmt.Printf("exp %v %v\n", exp.RuleName,exp.RuleResult)
	//}

	fmt.Println(pocResult.Vulnerable)


}

func TestCelParseYamlSets(t *testing.T) {
	setTypes     := make(map[string]ref.Val)

	// 解析target url
	parts, err := url.Parse("http://localhost:80")
	if err != nil {
		Log.Errorf("url parse error %v", err)
	}

	//根据schame host port 组合url
	//baseURL := fmt.Sprintf("%s://%s", parts.Scheme, parts.Host)
	//pathURL := fmt.Sprintf("%s://%s%s", parts.Scheme, parts.Host, URIPath(parts.Path))
	//host := parts.Hostname()

	defaultVars := map[string]interface{}{
		"request.url.scheme":   parts.Scheme,
		"request.url.domain":   parts.Hostname(),
		"request.url.host":     parts.Host,
		"request.url.port":     parts.Port(),
		"request.url.path":     parts.Path,
		"request.url.query":    parts.RawQuery,
		"request.url.fragment": parts.Fragment,
	}

	// 获取pocYaml
	pocYaml := GetPocYaml("poc-yaml-yapi-rce")


	//	set参数处理
	pocYaml.Sets= parseYamlSets(*pocYaml)

	celParseYamlSets(pocYaml.Sets,setTypes,defaultVars)

	//for k,v :=range setTypes{
	//	fmt.Println("k:",k)
	//	fmt.Println("v:",v)
	//}

}

func TestExpressionSlice(t *testing.T) {
	exp :="r0() && r1() && r2() && r3() && r4() && r5() && r6() && r7() || r8 || r9"
	sliceExp := expressionSlice(exp)
	for _,v :=range sliceExp.ExpResultSlice {
		fmt.Println(v.RuleName)
	}
}




















