package service

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
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

func (lm loggingMiddleware) Create(ctx context.Context, sd subscription.Data) (id string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Create(%v): %s, %s", sd, id, err))
	}()
	return lm.svc.Create(ctx, sd)
}

func (lm loggingMiddleware) Read(ctx context.Context, id string) (sd subscription.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Read(%s): (%v, %s)", id, sd, err))
	}()
	return lm.svc.Read(ctx, id)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s): %s", id, err))
	}()
	return lm.svc.Delete(ctx, id)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, q condition.Query, cursor string) (page []subscription.ConditionMatch, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchByCondition(ctx, q, cursor)
}

func (lm loggingMiddleware) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []subscription.Subscription, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByMetadata(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchByMetadata(ctx, q, cursor)
}
