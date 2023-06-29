package conditions_text

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
)

type serviceLogging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return serviceLogging{
		svc: svc,
		log: log,
	}
}

func (sl serviceLogging) Create(ctx context.Context, k, v string) (id string, err error) {
	id, err = sl.svc.Create(ctx, k, v)
	ll := sl.logLevel(err)
	sl.log.Log(ll, fmt.Sprintf("conditions_text.Create(k=%s, v=%s): id=%s, err=%s", k, v, id, err))
	return
}

func (sl serviceLogging) LockCreate(ctx context.Context, id string) (err error) {
	err = sl.svc.LockCreate(ctx, id)
	ll := sl.logLevel(err)
	sl.log.Log(ll, fmt.Sprintf("conditions_text.LockCreate(id=%s): err=%s", id, err))
	return
}

func (sl serviceLogging) UnlockCreate(ctx context.Context, id string) (err error) {
	err = sl.svc.UnlockCreate(ctx, id)
	ll := sl.logLevel(err)
	sl.log.Log(ll, fmt.Sprintf("conditions_text.UnlockCreate(id=%s): err=%s", id, err))
	return
}

func (sl serviceLogging) Delete(ctx context.Context, id string) (err error) {
	err = sl.svc.Delete(ctx, id)
	ll := sl.logLevel(err)
	sl.log.Log(ll, fmt.Sprintf("conditions_text.Delete(id=%s): err=%s", id, err))
	return
}

func (sl serviceLogging) logLevel(err error) (lvl slog.Level) {
	switch err {
	case nil:
		lvl = slog.DebugLevel
	default:
		lvl = slog.ErrorLevel
	}
	return
}
