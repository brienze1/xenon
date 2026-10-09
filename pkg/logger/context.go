package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Ctx interface {
	context.Context
	GetRequestId() string
	GetTransactionId() string
	GetCorrelationId() string
	SetCorrelationId(correlationId string)
}

type ctxKey struct{}

type logIds struct {
	requestId     string
	transactionId string
	correlationId string
	mu            sync.RWMutex
}

type logContext struct {
	context.Context
	ids *logIds
}

// GetContext returns the logger Ctx carried by the first context.Context argument, creating ids when absent.
func GetContext(ctx ...any) Ctx {
	parent := firstContext(ctx)
	if existing, ok := parent.(Ctx); ok {
		return existing
	}
	if ids, ok := parent.Value(ctxKey{}).(*logIds); ok {
		return &logContext{Context: parent, ids: ids}
	}
	transactionId := newId()
	ids := &logIds{requestId: newId(), transactionId: transactionId, correlationId: transactionId}
	return &logContext{Context: context.WithValue(parent, ctxKey{}, ids), ids: ids}
}

func (l *logContext) GetRequestId() string     { return l.ids.requestId }
func (l *logContext) GetTransactionId() string { return l.ids.transactionId }

func (l *logContext) GetCorrelationId() string {
	l.ids.mu.RLock()
	defer l.ids.mu.RUnlock()
	return l.ids.correlationId
}

func (l *logContext) SetCorrelationId(correlationId string) {
	l.ids.mu.Lock()
	defer l.ids.mu.Unlock()
	l.ids.correlationId = correlationId
}

func firstContext(candidates []any) context.Context {
	for _, candidate := range candidates {
		if base, ok := candidate.(context.Context); ok && base != nil {
			return base
		}
	}
	return context.Background()
}

func newId() string {
	buffer := make([]byte, 16)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
