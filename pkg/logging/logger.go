package logging

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var fallbackLogger = NewDefaultLogger().Named("fallback")

type loggerKey struct{}

func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
		return logger
	}
	return fallbackLogger
}

func NewLogger() (*zap.SugaredLogger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.OutputPaths = []string{"stdout"}
	config.DisableStacktrace = true
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}

func NewNamedFromConfig(cfg *zap.Config, name string) (*zap.SugaredLogger, error) {
	logger, err := NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return logger.Named(name), nil
}

func NewFromConfig(cfg *zap.Config) (*zap.SugaredLogger, error) {
	if cfg == nil {
		return NewDefaultLogger(), nil
	}
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("log level cannot be empty")
			}
		}()
		// this will panic if level is not set
		_ = cfg.Level.String()
		return
	}()
	if err != nil {
		return nil, err
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}

func NewDefaultLogger() *zap.SugaredLogger {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), os.Stdout, zap.DebugLevel)
	logger := zap.New(core).WithOptions(
		zap.AddCaller(),
		zap.AddStacktrace(zap.DPanicLevel),
		zap.Development(),
	)
	return logger.Sugar()
}
