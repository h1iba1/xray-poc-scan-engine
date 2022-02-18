package core

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"

)

// KV 键值对结构体
type KV struct {
	Key string
	Value string
}

type ProbesKV []KV

// 实现sort接口
func (ps ProbesKV) Len() int {
	return len(ps)
}

func (ps ProbesKV) Swap(i, j int) {
	ps[i],ps[j]=ps[j],ps[i]
}

func (ps ProbesKV) Less(i, j int) bool {
	return ps[i].Key<ps[j].Key
}

type PocYaml struct {
	Name string `yaml:"name"` // pocName

	Transport string `yaml:"transport"`	// 协议

	Set map[string]string `yaml:"set"`  // 设置的参数

	//建值对切片
	Sets ProbesKV `yaml:"-"`

	Rules map[string]Rules `yaml:"rules"` // rules r0 r1

	Expression string `yaml:"expression"` // rule执行规则 r0() && r1()

	Detail Detail `yaml:"detail"` // poc描述

	Outputs ProbesKV `yaml:"-"`
}

// Rn rule详细规则
type Rules struct {
	Request    Request `yaml:"request"`    //请求规则
	Expression string  `yaml:"expression"` //判断该条 Rule 的结果
	Output     Output  `yaml:"output"`     //声明一些变量，用于后续使用
}

// Request 请求规则
type Request struct {
	Vars ProbesKV `yaml:"-"`

	Cache bool `yaml:"cache"`
	Method string `yaml:"method"`
	Path string `yaml:"path"`
	Headers         map[string]string `yaml:"headers"`
	Body            string            `yaml:"body"`
	FollowRedirects bool              `yaml:"follow_redirects"`
}

// Output 请求输出结果
type Output struct {
	Search string `yaml:"search"`
	SearchReg    *regexp.Regexp `yaml:"-"`
	SearchOutput ProbesKV       `yaml:"-"`
}

// Detail poc基本信息
type Detail struct {
	Author      string   `yaml:"author"`
	Links       []string `yaml:"links"`


	Fingerprint Fingerprint `yaml:"fingerprint"` // 指纹信息

	Vulnerability Vulnerability `yaml:"vulnerability"` // 漏洞信息

	Summary ProbesKV `yaml:"-"` //其他未明确定义的字段
}

// 指纹信息
type Fingerprint struct {
	Infos    Infos    `yaml:"infos"` //指纹信息
	HostInfo HostInfo `yaml:"host_info"`
}

// 漏洞信息
type Vulnerability struct {
	Id string `yaml:"id"`		// "长亭漏洞库 id"
	Match string `yaml:"match"` // "证明漏洞存在的信息"
	Cve string `yaml:"cve"`
}

//指纹信息
type Infos struct {
	Id string `yaml:"id"`	// "长亭指纹库 id"
	Name string `yaml:"name"` //指纹name
	Version string `yaml:"version"`
	Type string `yaml:"type"`	// 指纹类型，有以下可选值： operating_system, hardware, system_bin, web_application, dependency
	Confidence string `yaml:"confidence"` //取值范围（1-100）
}

// 主机信息
type HostInfo struct {
	HostName string `yaml:"hostname"`	//主机名
}



func (p PocYaml) GetDetail() Detail {
	return p.Detail
}


// 将file数据转为pocYaml对象
func ParsePocByFile(file string)(*PocYaml,error){
	body,err :=ioutil.ReadFile(file)
	if err != nil {
		Log.Errorf("ParsePocByFile:%v",err)
	}
	return ParsePocByData(body)
}

// 将data []byte数据转为pocYaml
func ParsePocByData(data []byte) (*PocYaml,error) {
	result :=new(PocYaml)

	err :=yaml.Unmarshal(data,result)
	if err != nil {
		return nil,err
	}

	if result.Name==""{
		Log.Warningf("yaml name is empty:%v",string(data))
	}
	return result,nil
}