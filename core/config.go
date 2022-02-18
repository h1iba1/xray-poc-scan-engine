package core

import "xray-poc-scan-engine/logger"

var Log =logger.New()

// header头配置
type HeaderConfig struct {
	UserAgent []string `yaml:"UserAgent,flow"`
}

// http配置
type HttpConfig struct {
	DialTimeout     int          `yaml:"dial_timeout"`
	ReadTimeout     int          `yaml:"read_timeout"`
	FailRetries     int          `yaml:"fail_retries"`
	MaxQPS          int          `yaml:"max_qps"`
	MaxRedirect     int          `yaml:"max_redirect"`
	MaxConnsPerHost int          `yaml:"max_conns_per_host"`
	MaxRespBodySize int          `yaml:"max_resp_body_size"`
	Headers         HeaderConfig `yaml:"headers,flow"`
}


