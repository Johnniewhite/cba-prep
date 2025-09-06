package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level string, output string) (*Logger, error) {
	var config zap.Config

	if output == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	if output == "stdout" {
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	} else if output != "" {
		config.OutputPaths = []string{output}
		config.ErrorOutputPaths = []string{output}
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}, nil
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var args []interface{}
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		SugaredLogger: l.With(args...),
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		SugaredLogger: l.With("error", err),
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		SugaredLogger: l.With("request_id", requestID),
	}
}

func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{
		SugaredLogger: l.With("user_id", userID),
	}
}

func (l *Logger) Close() {
	_ = l.Sync()
}

func Default() *Logger {
	logger, _ := New("info", "stdout")
	return logger
}

func Fatal(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}