package logger

type LogoHandleFunc interface {
	Configure(*FileLogger) error
}

type OptionFn func(*FileLogger) error

func (o OptionFn) Configure (fl *FileLogger)error {
	return o(fl)
}
