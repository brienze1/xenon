package logger

import "context"

type Logger interface {
	Trace(ctx context.Context, message string, metadata ...any)
	Debug(ctx context.Context, message string, metadata ...any)
	Info(ctx context.Context, message string, metadata ...any)
	Warn(ctx context.Context, err error, message string, metadata ...any)
	Error(ctx context.Context, err error, message string, metadata ...any)
}

type loggerWrapper struct{}

func LoggerWrapper() Logger {
	return &loggerWrapper{}
}

func (l *loggerWrapper) Trace(ctx context.Context, message string, metadata ...any) {
	log(ctx, levelTrace, nil, message, metadata)
}

func (l *loggerWrapper) Debug(ctx context.Context, message string, metadata ...any) {
	log(ctx, levelDebug, nil, message, metadata)
}

func (l *loggerWrapper) Info(ctx context.Context, message string, metadata ...any) {
	log(ctx, levelInfo, nil, message, metadata)
}

func (l *loggerWrapper) Warn(ctx context.Context, err error, message string, metadata ...any) {
	log(ctx, levelWarn, err, message, metadata)
}

func (l *loggerWrapper) Error(ctx context.Context, err error, message string, metadata ...any) {
	log(ctx, levelError, err, message, metadata)
}
