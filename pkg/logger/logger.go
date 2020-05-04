package logger

import (
	"github.com/op/go-logging"
)

type NopBackend struct {
}

func (b *NopBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	// noop
	return nil
}

func (b *NopBackend) GetLevel(val string) logging.Level {
	return 0
}

func (b *NopBackend) SetLevel(level logging.Level, val string) {}
func (b *NopBackend) IsEnabledFor(level logging.Level, val string) bool {
	return false
}

type Logger struct {
	*logging.Logger
}

func (l Logger) Printf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l Logger) Println(args ...interface{}) {
	l.Info(args...)
}

func NewDummyLogger(name string) *Logger {
	gologger := logging.MustGetLogger(name)
	gologger.SetBackend(&NopBackend{})
	return &Logger{
		Logger: gologger,
	}
}

func NewLogger(name string) *Logger {
	gologger := logging.MustGetLogger(name)
	return &Logger{
		Logger: gologger,
	}
}
