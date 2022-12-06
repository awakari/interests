package matchers

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
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

func (lm loggingMiddleware) Create(ctx context.Context, k string, patternSrc string) (m model.MatcherData, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Create(%s, %s): %s", pkgName, k, patternSrc, err))
	}()
	return lm.svc.Create(ctx, k, patternSrc)
}

func (lm loggingMiddleware) LockCreate(ctx context.Context, patternCode model.PatternCode) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.LockCreate(%s): %s", pkgName, patternCode, err))
	}()
	return lm.svc.LockCreate(ctx, patternCode)
}

func (lm loggingMiddleware) UnlockCreate(ctx context.Context, patternCode model.PatternCode) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.UnlockCreate(%s): %s", pkgName, patternCode, err))
	}()
	return lm.svc.UnlockCreate(ctx, patternCode)
}

func (lm loggingMiddleware) Delete(ctx context.Context, m model.MatcherData) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Delete(%s): %s", pkgName, m, err))
	}()
	return lm.svc.Delete(ctx, m)
}
