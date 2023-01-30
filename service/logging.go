package service

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"golang.org/x/exp/slog"
)

type (
	loggingMiddleware struct {
		svc Service
		log *slog.Logger
	}
)

func NewLoggingMiddleware(svc Service, log *slog.Logger) Service {
	return loggingMiddleware{
		svc: svc,
		log: log,
	}
}

func (lm loggingMiddleware) Create(ctx context.Context, sub model.Subscription) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Create(%v): %s", sub, err))
	}()
	return lm.svc.Create(ctx, sub)
}

func (lm loggingMiddleware) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Read(%s): (%v, %s)", name, sub, err))
	}()
	return lm.svc.Read(ctx, name)
}

func (lm loggingMiddleware) Delete(ctx context.Context, name string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s): %s", name, err))
	}()
	return lm.svc.Delete(ctx, name)
}

func (lm loggingMiddleware) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("ListNames(%d, %s): %s", limit, cursor, err))
	}()
	return lm.svc.ListNames(ctx, limit, cursor)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, q model.ConditionQuery, cursor string) (page []model.Subscription, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchByCondition(ctx, q, cursor)
}
