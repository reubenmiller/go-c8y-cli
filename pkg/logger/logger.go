package logger

import (
	"os"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerInterface interface {
	Debugt(msg string, fields ...zapcore.Field)
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Debugs(args ...interface{})

	Infot(msg string, fields ...zapcore.Field)
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Infos(args ...interface{})

	Warnt(msg string, fields ...zapcore.Field)
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Warns(args ...interface{})

	Errort(msg string, fields ...zapcore.Field)
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Errors(args ...interface{})

	Panict(msg string, fields ...zapcore.Field)
	Panicf(template string, args ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Panic(msg string, keysAndValues ...interface{})
	Panics(args ...interface{})

	Fatalt(msg string, fields ...zapcore.Field)
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	Fatals(args ...interface{})

	AtLevel(level zapcore.Level, msg string, fields ...zapcore.Field) *Logger
}

// Logger provides a log interface to verbose messages to the user
type Logger struct {
	zlogger *zap.Logger
}

// Printf is an alias for Infof
func (l Logger) Printf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

// Println is an alias for Info
func (l Logger) Println(args ...interface{}) {
	l.Info(args...)
}

// Warningf logs a warning message with a format string
func (l Logger) Warningf(format string, args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Warnf(format, args...)
	}
}

// Warning logs a warning message
func (l Logger) Warning(args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Warn(args...)
	}
}

// Errorf logs an error message with a format string
func (l Logger) Errorf(format string, args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Errorf(format, args...)
	}
}

// Error logs an error message
func (l Logger) Error(args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Error(args...)
	}
}

// Debug logs a debug message
func (l Logger) Debug(args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Debug(args...)
	}
}

// Debugf logs a debug message with a format string
func (l Logger) Debugf(format string, args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Debugf(format, args...)
	}
}

// Info logs a information message
func (l Logger) Info(args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Info(args...)
	}
}

// Infof logs a information message with a format string
func (l Logger) Infof(format string, args ...interface{}) {
	if l.zlogger != nil {
		l.zlogger.Sugar().Infof(format, args...)
	}
}

// NewDummyLogger create a dummy no-op logger
func NewDummyLogger(name string) *Logger {
	return NewLogger(name, Options{
		Silent: true,
	})
}

// Options holds settings of the logger
type Options struct {
	// Silent removes silences all log messages
	Silent bool

	// Color prints the log levels in color
	Color bool

	// Debug activates all log messages
	Debug bool
}

// NewLogger create a logger with given options
func NewLogger(name string, options Options) *Logger {
	// new logger
	consoleEncCfg := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleLevel := zapcore.InfoLevel
	if options.Debug {
		consoleLevel = zapcore.DebugLevel
	}
	consoleLevelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= consoleLevel
	})

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncCfg)

	var cores []zapcore.Core

	if options.Silent {
		cores = append(cores, zapcore.NewNopCore())
	} else {
		output := colorable.NewNonColorable(os.Stderr)
		if options.Color {
			output = colorable.NewColorableStderr()
		}
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(output), consoleLevelEnabler))
	}

	core := zapcore.NewTee(cores...)
	unsugared := zap.New(core)

	return &Logger{
		zlogger: unsugared,
	}
}
