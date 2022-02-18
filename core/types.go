package core

import (
	"net/http"
	"time"
)

// poc请求
// 所依赖的参数
type PocRequest struct {
	URL         string            `json:"url"`
	YmlName     string            `json:"yml_name"`
	Headers     http.Header       `json:"headers"`
	Method      string            `json:"method"`
	PostData    string            `json:"post_data"`
	Port        int               `json:"port"`
	//Options     map[string]string `json:"options"`     //poc利用用来保存参数
	//PreOptions  map[string]string `json:"pre_options"` //poc依赖用来保存参数
	//PreResponse *PreResponse      `json:"pre_response"`
}

//type PreResponse struct {
//	StatusCode  int64
//	Headers     http.Header
//	Body        string
//	ResponseUrl string
//}

// Poc执行信息
type PocResult struct {
	Vulnerable bool   `json:"vulnerable"`
	Detail     Detail `json:"detail"` // 详细信息

	Target    PocTarget `json:"target"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	//RequestCount int64                  `json:"request_count"`
	Outputs      map[string]string      `json:"outputs"`
	PrePoc       bool                   `json:"pre_poc"` //前置poc
	IsVerify     bool                   `json:"is_verify"`

	ExpResult  ExpResult
	Expression string
	PocInfo    PocInfo
}

type ExpResult struct {
	ExpResultSlice []RuleResultJudge
}

type RuleResultJudge struct {
	RuleName string
	RuleResult bool
}

// 插件信息
type PocInfo struct {
	PocName     string
	PocLevel    int
}

type PocTarget struct {
	Typ   string     `json:"typ"` // 扫描目标类型 [web/system]
	URL   string     `json:"url"`
	Host  string     `json:"host"`
	Port  int        `json:"port"`
	Param []PocParam `json:"params"`
}

type PocParam struct {
	Method string      `json:"method"`
	Param  string      `json:"param"`
	Verify []PocVerify `json:"verify"`
}

type PocVerify struct {
	Payload string  `json:"payload"`
	HttpRaw HttpRaw `json:"http_raw"`
}

type HttpRaw struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}