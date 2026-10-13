package logger

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"time"
)

type logPattern struct {
	Exceptions    any    `json:"exceptions,omitempty"`
	RequestId     string `json:"requestId,omitempty"`
	Timestamp     string `json:"timestamp"`
	Level         string `json:"level"`
	Service       string `json:"service"`
	Operation     string `json:"operation"`
	Detail        string `json:"detail"`
	CorrelationId string `json:"correlationId,omitempty"`
	TransactionId string `json:"transactionId,omitempty"`
	Metadata      []any  `json:"metadata,omitempty"`
}

func Error(ctx context.Context, err error, message string, metadata ...any) {
	log(ctx, levelError, err, message, metadata)
}

func Warn(ctx context.Context, err error, message string, metadata ...any) {
	log(ctx, levelWarn, err, message, metadata)
}

func Info(ctx context.Context, message string, metadata ...any) {
	log(ctx, levelInfo, nil, message, metadata)
}

func Debug(ctx context.Context, message string, metadata ...any) {
	log(ctx, levelDebug, nil, message, metadata)
}

func Trace(ctx context.Context, message string, metadata ...any) {
	log(ctx, levelTrace, nil, message, metadata)
}

func log(ctx context.Context, logLevel level, err error, message string, metadata []any) {
	if !shouldPrintLog(logLevel) {
		return
	}
	pattern := logPattern{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     logLevel.String(),
		Service:   serviceName(),
		Operation: operationName(),
		Detail:    message,
		Metadata:  redact(metadata),
	}
	if model := newErrorModel(err); model != nil {
		pattern.Exceptions = model
	}
	if ctx != nil {
		logCtx := GetContext(ctx)
		pattern.RequestId = logCtx.GetRequestId()
		pattern.TransactionId = logCtx.GetTransactionId()
		pattern.CorrelationId = logCtx.GetCorrelationId()
	}
	line, marshalErr := json.Marshal(pattern)
	if marshalErr != nil {
		pattern.Metadata = []any{"unserializable metadata"}
		line, _ = json.Marshal(pattern)
	}
	write(append(redactLine(ctx, line), '\n'))
}

func operationName() string {
	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	function := runtime.FuncForPC(pc)
	if function == nil {
		return ""
	}
	name := function.Name()
	return name[strings.LastIndex(name, "/")+1:]
}
