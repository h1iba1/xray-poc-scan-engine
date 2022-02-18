package utils

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
// RandomLowercase 指定长度的小写字母组成的随机字符串
func RandomLowercase(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// HttpOptions http设置
type HttpOptions struct {
	DialTimeout     int
	ReadTimeout     int
	Proxy           string
	Jar             *cookiejar.Jar
	MaxRedirect     int
	MaxConnsPerHost int
	MaxQPS          int
	NeedRedirect    bool
}

// NewHttpClient 生成新的httpClient
func NewHttpClient(options *HttpOptions) (*http.Client, error) {
	// 超时时间
	client := &http.Client{
		Timeout: time.Duration(options.ReadTimeout) * time.Second,
	}
	// cookie
	if options.Jar != nil {
		client.Jar = options.Jar
	}

	// http拨号客户端设置
	tran := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
		MaxIdleConns:    options.MaxQPS,
		MaxConnsPerHost: options.MaxConnsPerHost,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(options.DialTimeout) * time.Second,
			KeepAlive: time.Duration(options.ReadTimeout) * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// proxy设置

	// 客户端Transport等于设置过代理的tran
	client.Transport = tran
	// 是否需要重定向
	if options.NeedRedirect == false {
		client.CheckRedirect = noRedirect
	} else {
		if options.MaxRedirect > 0 {
			client.CheckRedirect = defaultLimitRedirect(options.MaxRedirect)
		}
	}

	return client, nil
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// 限制重定向数
func defaultLimitRedirect(maxRedirect int) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) > maxRedirect {
			return fmt.Errorf("stopped after %d redirects", maxRedirect)
		}
		return nil
	}
}

// URIPath 解析uri 获取所有path
func URIPath(uri string) string {
	if uri == "" {
		return ""
	}
	parts := strings.Split(uri, "/")
	l := len(parts)
	last := parts[l-1]
	// 包含. 说明为./test一类的相对路径
	if strings.Contains(last, ".") {
		l = l - 1
	}
	var res string
	// 循环组合/path
	for i := 0; i < l; i++ {
		if parts[i] == "" {
			continue
		}
		res += "/" + parts[i]
	}
	return res
}

// GetHttpPort 获取url port
func GetHttpPort(parts *url.URL) int {
	var port int
	isHTTPS := parts.Scheme == "https"
	portStr := parts.Port()
	if portStr == "" {
		if isHTTPS {
			port = 443
		} else {
			port = 80
		}
	} else {
		intValue, err := strconv.Atoi(portStr)
		if err != nil {
			return 0
		}
		port = intValue
	}
	return port
}


