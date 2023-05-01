package service

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/util"
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

func (lm loggingMiddleware) Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Create(%s, %s, %+v): %s, %s", groupId, userId, sd, id, err))
	}()
	return lm.svc.Create(ctx, groupId, userId, sd)
}

func (lm loggingMiddleware) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Read(%s, %s, %s): (%v, %s)", id, groupId, userId, sd, err))
	}()
	return lm.svc.Read(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) UpdateMetadata(ctx context.Context, id, groupId, userId string, md subscription.Metadata) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("UpdateMetadata(%s, %s, %s, %+v): err=%s", id, groupId, userId, md, err))
	}()
	return lm.svc.UpdateMetadata(ctx, id, groupId, userId, md)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id, groupId, userId string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s, %s, %s): %s", id, groupId, userId, err))
	}()
	return lm.svc.Delete(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, cond condition.Condition, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(%+v): %s", cond, err))
	}()
	return lm.svc.SearchByCondition(ctx, cond, consumeFunc)
}

func (lm loggingMiddleware) SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByAccount(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchByAccount(ctx, q, cursor)
}
