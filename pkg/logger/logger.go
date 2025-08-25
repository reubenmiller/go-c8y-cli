package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// Logger provides a log interface to verbose messages to the user
// type Logger struct {
// 	zLogger *slog.Logger
// }

// // Warn logs a warning message
// func (l Logger) Warn(msg string, args ...interface{}) {
// 	if l.zLogger != nil {
// 		l.zLogger.Warn(msg, args...)
// 	}
// }

// // Error logs an error message
// func (l Logger) Error(msg string, args ...interface{}) {
// 	if l.zLogger != nil {
// 		l.zLogger.Error(msg, args...)
// 	}
// }

// // Debug logs a debug message
// func (l Logger) Debug(msg string, args ...interface{}) {
// 	if l.zLogger != nil {
// 		l.zLogger.Debug(msg, args...)
// 	}
// }

// // Info logs a information message
// func (l Logger) Info(msg string, args ...interface{}) {
// 	if l.zLogger != nil {
// 		l.zLogger.Info(msg, args...)
// 	}
// }

// NewDummyLogger create a dummy no-op logger
func NewDummyLogger(name string) *slog.Logger {
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

	// Level default log level
	Level slog.Level
}

// NewLogger create a logger with given options
func NewLogger(name string, options Options) *slog.Logger {
	// new logger
	lvl := new(slog.LevelVar)

	var logger *slog.Logger
	w := os.Stderr
	if options.Silent {
		logger = slog.New(slog.DiscardHandler)
	} else {
		logger = slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      lvl,
				TimeFormat: time.RFC3339,
				NoColor:    !options.Color,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Value.Kind() == slog.KindAny {
						if _, ok := a.Value.Any().(error); ok {
							return tint.Attr(9, a)
						}
					}
					return a
				},
			}),
		)
	}

	lvl.Set(options.Level)
	if options.Debug {
		lvl.Set(slog.LevelDebug)
	} else {
		lvl.Set(options.Level)
	}
	slog.SetDefault(logger)
	return logger
}
