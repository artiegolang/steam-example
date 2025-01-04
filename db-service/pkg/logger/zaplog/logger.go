package zaplog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"petProject/db-service/config"
	"petProject/db-service/pkg/logger"
	"sync"
)

type ZapLogger struct {
	log *zap.Logger
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

var (
	globalLogger logger.Logger
	once         sync.Once
)

func InitLogger(cfg config.LogConfig) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      cfg.Development,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	globalLogger = &ZapLogger{log: logger}
}

func GetLogger() logger.Logger {
	if globalLogger == nil {
		panic("logger is not initialized")
	}
	return globalLogger
}
