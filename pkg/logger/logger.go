package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger представляет интерфейс для логирования
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger создает новый экземпляр логгера
func NewLogger(level string) (*Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       false,
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     false,
		DisableStacktrace: false,
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации zap логгера: %w", err)
	}

	sugar := logger.Sugar()
	return &Logger{sugar}, nil
}

// Fatal логирует сообщение с уровнем Fatal и завершает программу
func (l *Logger) Fatal(msg string, err error) {
	l.Fatalw(msg, "error", err)
}

// Error логирует сообщение с уровнем Error
func (l *Logger) Error(msg string, err error) {
	l.Errorw(msg, "error", err)
}

// Info логирует сообщение с уровнем Info
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Infow(msg, args...)
}

// Debug логирует сообщение с уровнем Debug
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Debugw(msg, args...)
}

// Warn логирует сообщение с уровнем Warn
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Warnw(msg, args...)
}
