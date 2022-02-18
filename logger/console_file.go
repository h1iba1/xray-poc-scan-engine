package logger

import (
	"os"
	"time"
)

type ConsoleFileLogger struct {
	consoleLogger *ConsoleLogger
	fileLogger    *FileLogger
}

func NewConsoleFileLogger(stdout LogLevel, file LogLevel, fp, fn string, options ...LogoHandleFunc) *ConsoleFileLogger {
	cl := NewConsoleLogger(stdout)
	fl := NewFileLogger(file, fp, fn, options...)
	return &ConsoleFileLogger{consoleLogger: cl, fileLogger: fl}
}

func (cf *ConsoleFileLogger) SetPrefix(prefix string) {
	cf.consoleLogger.Prefix = prefix
	cf.fileLogger.Prefix = prefix
}

func (cf *ConsoleFileLogger) SetTimeFormat(format string) {
	logTimeFormat = format
}

func (cf *ConsoleFileLogger) SetFileInfo(level LogLevel) {
	cf.consoleLogger.FileInfoLevel = level
	cf.fileLogger.FileInfoLevel = level
}

func (cf *ConsoleFileLogger) SetEncoder(encode encoder)  {
	cf.consoleLogger.Encoder = encode
	cf.fileLogger.Encoder = encode
}

func (cf *ConsoleFileLogger) Debug(msg string) {
	cf.consoleLogger.Log(DebugLevel, msg)
	cf.fileLogger.Log(DebugLevel, msg)
}

func (cf *ConsoleFileLogger) Trace(msg string) {
	cf.consoleLogger.Log(TraceLevel, msg)
	cf.fileLogger.Log(TraceLevel, msg)
}

func (cf *ConsoleFileLogger) Info(msg string) {
	cf.consoleLogger.Log(InfoLevel, msg)
	cf.fileLogger.Log(InfoLevel, msg)
}

func (cf *ConsoleFileLogger) Warning(msg string) {
	cf.consoleLogger.Log(WarningLevel, msg)
	cf.fileLogger.Log(WarningLevel, msg)
}

func (cf *ConsoleFileLogger) Error(msg string) {
	cf.consoleLogger.Log(ErrorLevel, msg)
	cf.fileLogger.Log(ErrorLevel, msg)
}

func (cf *ConsoleFileLogger) Fatal(msg string) {
	cf.consoleLogger.Log(FatalLevel, msg)
	cf.fileLogger.Log(FatalLevel, msg)
	time.Sleep(time.Second)
	os.Exit(1)
}

func (cf *ConsoleFileLogger) Debugf(format string, a ...interface{}) {
	cf.consoleLogger.Log(DebugLevel, format, a...)
	cf.fileLogger.Log(DebugLevel, format, a...)
}

func (cf *ConsoleFileLogger) Tracef(format string, a ...interface{}) {
	cf.consoleLogger.Log(TraceLevel, format, a...)
	cf.fileLogger.Log(TraceLevel, format, a...)
}

func (cf *ConsoleFileLogger) Infof(format string, a ...interface{}) {
	cf.consoleLogger.Log(InfoLevel, format, a...)
	cf.fileLogger.Log(InfoLevel, format, a...)
}

func (cf *ConsoleFileLogger) Warningf(format string, a ...interface{}) {
	cf.consoleLogger.Log(WarningLevel, format, a...)
	cf.fileLogger.Log(WarningLevel, format, a...)
}

func (cf *ConsoleFileLogger) Errorf(format string, a ...interface{}) {
	cf.consoleLogger.Log(ErrorLevel, format, a...)
	cf.fileLogger.Log(ErrorLevel, format, a...)
}

func (cf *ConsoleFileLogger) Fatalf(format string, a ...interface{}) {
	cf.consoleLogger.Log(FatalLevel, format, a...)
	cf.fileLogger.Log(FatalLevel, format, a...)
	time.Sleep(time.Second)
	os.Exit(1)
}
