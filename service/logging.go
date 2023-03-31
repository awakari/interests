package service

import (
	"context"
	"fmt"
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

func (lm loggingMiddleware) Create(ctx context.Context, acc string, sd subscription.Data) (id string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Create(%s, %+v): %s, %s", acc, sd, id, err))
	}()
	return lm.svc.Create(ctx, acc, sd)
}

func (lm loggingMiddleware) Read(ctx context.Context, id, acc string) (sd subscription.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Read(%s, %s): (%v, %s)", id, acc, sd, err))
	}()
	return lm.svc.Read(ctx, id, acc)
}

func (lm loggingMiddleware) UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("UpdateMetadata(%s, %s, %+v): err=%s", id, acc, md, err))
	}()
	return lm.svc.UpdateMetadata(ctx, id, acc, md)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id, acc string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s, %s): %s", id, acc, err))
	}()
	return lm.svc.Delete(ctx, id, acc)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, q condition.Query, cursor subscription.ConditionMatchKey) (page []subscription.ConditionMatch, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchByCondition(ctx, q, cursor)
}

func (lm loggingMiddleware) SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByAccount(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchByAccount(ctx, q, cursor)
}
