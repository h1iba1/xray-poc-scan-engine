package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

type ConsoleLogger struct {
	Level        LogLevel `json:"level"`
	Prefix       string   `json:"prefix"`
	FileInfoLevel LogLevel `json:"file_info_level"`
	Encoder      encoder  `json:"encoder"`
}

// ConsoleLogger 构造函数
func NewConsoleLogger(level LogLevel, args ...interface{}) *ConsoleLogger {
	//level, err := parseLogLevel(Level)
	//if err != nil {
	//	panic(err)
	//}
	consoleLogger := &ConsoleLogger{Level: level, FileInfoLevel:FatalLevel}
	for _, arg := range args {
		switch arg.(type) {
		case encoder:
			encode := arg.(encoder)
			consoleLogger.Encoder = encode
		}
	}
	return consoleLogger
}

func (cl *ConsoleLogger) Log(level LogLevel, format string, args ...interface{}) {
	if cl.Level <= level {
		var content []byte
		var msg string
		if len(args) > 0 {
			msg = fmt.Sprintf(format, args...)
		} else {
			msg = format
		}
		now := time.Now()
		if cl.FileInfoLevel <= level {
			info, err := getInfo(3)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if cl.Encoder == JsonEncoder {
				content = cl.JsonEncode(now.Format(logTimeFormat), info, level.String(), msg)
			} else {
				content = cl.TextEncode(now.Format(logTimeFormat), info, level.String(), msg)
			}
			_, _ = io.Writer.Write(os.Stdout, content)
			return
		}
		if cl.Encoder == JsonEncoder {
			content = cl.JsonEncode(now.Format(logTimeFormat), "", level.String(), msg)
		} else {
			content = cl.TextEncode(now.Format(logTimeFormat), "", level.String(), msg)
		}
		_, _ = io.Writer.Write(os.Stdout, content)
		return
	}
}

func (cl *ConsoleLogger) TextEncode(formatTime, fileInfo, level, msg string) []byte {
	buf := bytes.Buffer{}
	if cl.Prefix != "" {
		buf.WriteString(cl.Prefix)
		buf.WriteString(" ")
	}
	buf.WriteString(formatTime)
	buf.WriteString(" ")
	if fileInfo != "" {
		buf.WriteString(fileInfo)
	}
	buf.WriteString(" ▶ [")
	buf.WriteString(level)
	buf.WriteString("] ")
	buf.WriteString(msg)
	buf.WriteString("\n")
	return buf.Bytes()
}

func (cl *ConsoleLogger) JsonEncode(formatTime, fileInfo, level, msg string) []byte {
	buf := bytes.Buffer{}
	buf.WriteString(`{`)
	if cl.Prefix != "" {
		buf.WriteString(`"prefix": "`)
		buf.WriteString(cl.Prefix)
		buf.WriteString(`",`)
	}
	buf.WriteString(`"time": "`)
	buf.WriteString(formatTime)
	if fileInfo != "" {
		buf.WriteString(`","fileInfo": "`)
		buf.WriteString(fileInfo)
	}
	buf.WriteString(`","level": "`)
	buf.WriteString(level)
	buf.WriteString(`","msg": "`)
	buf.WriteString(msg)
	buf.WriteString("\"}\n")
	return buf.Bytes()
}

func (cl *ConsoleLogger) SetPrefix(prefix string) {
	cl.Prefix = prefix
}

func (cl *ConsoleLogger) SetTimeFormat(format string) {
	logTimeFormat = format
}

func (cl *ConsoleLogger) SetFileInfo(level LogLevel) {
	cl.FileInfoLevel= level
}

func (cl *ConsoleLogger) SetEncoder(encode encoder) {
	cl.Encoder = encode
}

func (cl *ConsoleLogger) Debug(msg string) {
	cl.Log(DebugLevel, msg)
}

func (cl *ConsoleLogger) Trace(msg string) {
	cl.Log(TraceLevel, msg)
}

func (cl *ConsoleLogger) Info(msg string) {
	cl.Log(InfoLevel, msg)
}

func (cl *ConsoleLogger) Warning(msg string) {
	cl.Log(WarningLevel, msg)
}

func (cl *ConsoleLogger) Error(msg string) {
	cl.Log(ErrorLevel, msg)
}

func (cl *ConsoleLogger) Fatal(msg string) {
	cl.Log(FatalLevel, msg)
	os.Exit(1)
}

func (cl *ConsoleLogger) Debugf(format string, args ...interface{}) {
	cl.Log(DebugLevel, format, args...)
}

func (cl *ConsoleLogger) Tracef(format string, args ...interface{}) {
	cl.Log(TraceLevel, format, args...)
}

func (cl *ConsoleLogger) Infof(format string, args ...interface{}) {
	cl.Log(InfoLevel, format, args...)
}

func (cl *ConsoleLogger) Warningf(format string, args ...interface{}) {
	cl.Log(WarningLevel, format, args...)
}

func (cl *ConsoleLogger) Errorf(format string, args ...interface{}) {
	cl.Log(ErrorLevel, format, args...)
}

func (cl *ConsoleLogger) Fatalf(format string, args ...interface{}) {
	cl.Log(FatalLevel, format, args...)
	os.Exit(1)
}
