package logger

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type LogLevel uint8
type encoder uint8

type Logger interface {
	Debug(msg string)
	Trace(msg string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	Fatal(msg string)
	Debugf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	SetPrefix(prefix string)
	SetTimeFormat(format string)
	SetFileInfo(level LogLevel)
	SetEncoder(encode encoder)
}

type MaxFileCount struct {
	FileCount int
	ErrFileCount int
}

//type MaxFileSize struct {
//	Size int64
//}

//type FileAge struct {
//	SplitFileAge int
//	MaxFileAge int
//}

const (
	UnKnow LogLevel = iota
	DebugLevel
	TraceLevel
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

const (
	TextEncoder encoder = iota
	JsonEncoder
)

var (
	levels = map[LogLevel]string{
		DebugLevel: "Debug",
		TraceLevel: "Trace",
		InfoLevel:  "Info",
		WarningLevel:  "Warning",
		ErrorLevel: "Error",
		FatalLevel: "Fatal",
	}
	// 日志时间格式字符串
	logTimeFormat = "2006/01/02 - 15:04:05.000"
	// 是否打印文件行号信息 吗，偶人为true
	maxChanSize = 50000
)

func (ll LogLevel) String () string {
	return levels[ll]
}

func ParseLogLevel(s string) (LogLevel, error)  {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	case "info":
		return InfoLevel, nil
	case "warning":
		return WarningLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return UnKnow, errors.New("无效的日志级别")
	}
}

func getInfo(n int) (string, error)  {
	pc, fileName, lineNo, ok := runtime.Caller(n)
	if !ok {
		return "", errors.New("runtime.Caller() failed\n")
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[1]
	//return fmt.Sprintf("%s:%s:%d", fileName, funcName, lineNo), nil
	return fmt.Sprintf("%s:%d", fileName, lineNo), nil
}

func New() *ConsoleLogger {
	return NewConsoleLogger(DebugLevel)
}