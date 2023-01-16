package kiwiTree

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
)

type (
	loggingMiddleware struct {
		svc Service
		log *slog.Logger
	}
)

const (
	pkgName = "matchers"
)

func NewLoggingMiddleware(svc Service, log *slog.Logger) Service {
	return loggingMiddleware{
		svc: svc,
		log: log,
	}
}

func (lm loggingMiddleware) Create(ctx context.Context, k string, pattern string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Create(%s, %s): %s", pkgName, k, pattern, err))
	}()
	return lm.svc.Create(ctx, k, pattern)
}

func (lm loggingMiddleware) LockCreate(ctx context.Context, k string, pattern string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.LockCreate(%s, %s): %s", pkgName, k, pattern, err))
	}()
	return lm.svc.LockCreate(ctx, k, pattern)
}

func (lm loggingMiddleware) UnlockCreate(ctx context.Context, k string, pattern string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.UnlockCreate(%s, %s): %s", pkgName, k, pattern, err))
	}()
	return lm.svc.UnlockCreate(ctx, k, pattern)
}

func (lm loggingMiddleware) Delete(ctx context.Context, k string, pattern string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Delete(%s, %s): %s", pkgName, k, pattern, err))
	}()
	return lm.svc.Delete(ctx, k, pattern)
}
