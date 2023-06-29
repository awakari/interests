package service

import (
	"context"
	"fmt"
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

func (lm loggingMiddleware) Update(ctx context.Context, id, groupId, userId string, d subscription.Data) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Update(%s, %s, %s, %+v): err=%s", id, groupId, userId, d, err))
	}()
	return lm.svc.Update(ctx, id, groupId, userId, d)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id, groupId, userId string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s, %s, %s): %s", id, groupId, userId, err))
	}()
	return lm.svc.Delete(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchOwn(%v, %v): %s", q, cursor, err))
	}()
	return lm.svc.SearchOwn(ctx, q, cursor)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, condId string, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(%s): %s", condId, err))
	}()
	return lm.svc.SearchByCondition(ctx, condId, consumeFunc)
}
