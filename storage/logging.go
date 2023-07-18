package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model/subscription"
	"golang.org/x/exp/slog"
)

type loggingMiddleware struct {
	stor Storage
	log  *slog.Logger
}

func NewLoggingMiddleware(stor Storage, log *slog.Logger) Storage {
	return loggingMiddleware{
		stor: stor,
		log:  log,
	}
}

func (lm loggingMiddleware) Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Create(%s, %s, %+v): %s, %s", groupId, userId, sd, id, err))
	}()
	return lm.stor.Create(ctx, groupId, userId, sd)
}

func (lm loggingMiddleware) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Read(%s, %s, %s): (%v, %s)", id, groupId, userId, sd, err))
	}()
	return lm.stor.Read(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) Update(ctx context.Context, id, groupId, userId string, d subscription.Data) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Update(%s, %s, %s, %+v): err=%s", id, groupId, userId, d, err))
	}()
	return lm.stor.Update(ctx, id, groupId, userId, d)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id, groupId, userId string) (d subscription.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s, %s, %s): %s", id, groupId, userId, err))
	}()
	return lm.stor.Delete(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchOwn(%v, %v): %s", q, cursor, err))
	}()
	return lm.stor.SearchOwn(ctx, q, cursor)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, q subscription.QueryByCondition, cursor string) (page []subscription.ConditionMatch, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(q=%+v, cursor=%s): %d, %s", q, cursor, len(page), err))
	}()
	return lm.stor.SearchByCondition(ctx, q, cursor)
}

func (lm loggingMiddleware) Close() (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Close(): %s", err))
	}()
	return lm.stor.Close()
}
