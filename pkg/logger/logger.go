package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger represents the logging interface
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level string) (*Logger, error) {
	config := zap.NewProductionConfig()

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Configure output format
	config.Encoding = "json"
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Create logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// Fatal logs a message with Fatal level and exits the program
func (l *Logger) Fatal(msg string, err error) {
	l.Logger.Fatal(msg, zap.Error(err))
}

// Error logs a message with Error level
func (l *Logger) Error(msg string, err error) {
	l.Logger.Error(msg, zap.Error(err))
}

// Info logs a message with Info level
func (l *Logger) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		var zapFields []zap.Field
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key, ok := fields[i].(string)
				if ok {
					zapFields = append(zapFields, zap.Any(key, fields[i+1]))
				}
			}
		}
		l.Logger.Info(msg, zapFields...)
	} else {
		l.Logger.Info(msg)
	}
}

// Debug logs a message with Debug level
func (l *Logger) Debug(msg string) {
	l.Logger.Debug(msg)
}

// Warn logs a message with Warn level
func (l *Logger) Warn(msg string) {
	l.Logger.Warn(msg)
}
