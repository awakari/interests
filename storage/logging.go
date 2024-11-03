package storage

import (
	"context"
	"fmt"
	"github.com/awakari/interests/model/interest"
	"log/slog"
	"time"
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

func (lm loggingMiddleware) Create(ctx context.Context, id, groupId, userId string, sd interest.Data) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Create(%s, %s, %s, %+v): %s", id, groupId, userId, sd, err))
	}()
	return lm.stor.Create(ctx, id, groupId, userId, sd)
}

func (lm loggingMiddleware) Read(ctx context.Context, id, groupId, userId string) (sd interest.Data, ownerGroupId, ownerUserId string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Read(%s, %s, %s): (%v, %s, %s, %s)", id, groupId, userId, sd, ownerGroupId, ownerUserId, err))
	}()
	return lm.stor.Read(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) Update(ctx context.Context, id, groupId, userId string, d interest.Data) (prev interest.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Update(%s, %s, %s, %+v): err=%s", id, groupId, userId, d, err))
	}()
	return lm.stor.Update(ctx, id, groupId, userId, d)
}

func (lm loggingMiddleware) UpdateFollowers(ctx context.Context, id string, count int64) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("UpdateFollowers(%s, %d): err=%s", id, count, err))
	}()
	return lm.stor.UpdateFollowers(ctx, id, count)
}

func (lm loggingMiddleware) UpdateResultTime(ctx context.Context, id string, last time.Time) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("UpdateResultTime(%s, %s): err=%s", id, last, err))
	}()
	return lm.stor.UpdateResultTime(ctx, id, last)
}

func (lm loggingMiddleware) SetEnabledBatch(ctx context.Context, ids []string, enabled bool) (n int64, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SetEnabledBatch(%+v, %t): %d, err=%s", ids, enabled, n, err))
	}()
	return lm.stor.SetEnabledBatch(ctx, ids, enabled)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id, groupId, userId string) (d interest.Data, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Delete(%s, %s, %s): %s", id, groupId, userId, err))
	}()
	return lm.stor.Delete(ctx, id, groupId, userId)
}

func (lm loggingMiddleware) Search(ctx context.Context, q interest.Query, cursor interest.Cursor) (ids []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Search(%v, %v): %d, %s", q, cursor, len(ids), err))
	}()
	return lm.stor.Search(ctx, q, cursor)
}

func (lm loggingMiddleware) SearchByCondition(ctx context.Context, q interest.QueryByCondition, cursor string) (page []interest.ConditionMatch, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("SearchByCondition(q=%+v, cursor=%s): %d, %s", q, cursor, len(page), err))
	}()
	return lm.stor.SearchByCondition(ctx, q, cursor)
}

func (lm loggingMiddleware) Count(ctx context.Context) (count int64, err error) {
	count, err = lm.stor.Count(ctx)
	lm.log.Debug(fmt.Sprintf("Count(): %d, %s", count, err))
	return
}

func (lm loggingMiddleware) CountUsersUnique(ctx context.Context) (count int64, err error) {
	count, err = lm.stor.CountUsersUnique(ctx)
	lm.log.Debug(fmt.Sprintf("CountUsersUnique(): %d, %s", count, err))
	return
}

func (lm loggingMiddleware) Close() (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("Close(): %s", err))
	}()
	return lm.stor.Close()
}
