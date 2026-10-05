package logx

import (
	"context"

	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// New creates a new Logger with the specified configuration.
func New(cfg *LoggerConfig) (*Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Production {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar()

	return &Logger{
		logger: logger,
		sugar:  sugar,
	}, nil
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...any) {
	l.sugar.With(ContextFields(ctx)...).Debugw(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...any) {
	l.sugar.With(ContextFields(ctx)...).Infow(msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...any) {
	l.sugar.With(ContextFields(ctx)...).Warnw(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...any) {
	l.sugar.With(ContextFields(ctx)...).Errorw(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...any) {
	l.sugar.With(ContextFields(ctx)...).Fatalw(msg, fields...)
}

func (l *Logger) Panic(ctx context.Context, msg string, fields ...any) {
	l.sugar.With(ContextFields(ctx)...).Panicw(msg, fields...)
}
