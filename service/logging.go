package service

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/sirupsen/logrus"
	"reflect"
)

type (
	loggingMiddleware struct {
		svc Service
		log *logrus.Entry
	}
)

var (
	pkgName = reflect.TypeOf(loggingMiddleware{}).PkgPath()
)

func NewLoggingMiddleware(svc Service, log *logrus.Entry) Service {
	return loggingMiddleware{
		svc: svc,
		log: log,
	}
}

func (lm loggingMiddleware) Create(ctx context.Context, name string, req CreateRequest) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Create(%s, %v): %s", pkgName, name, req, err))
	}()
	return lm.svc.Create(ctx, name, req)
}

func (lm loggingMiddleware) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Read(%s): (%v, %s)", pkgName, name, sub, err))
	}()
	return lm.svc.Read(ctx, name)
}

func (lm loggingMiddleware) Delete(ctx context.Context, name string) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Delete(%s): %s", pkgName, name, err))
	}()
	return lm.svc.Delete(ctx, name)
}

func (lm loggingMiddleware) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.ListNames(%d, %s): %s", pkgName, limit, cursor, err))
	}()
	return lm.svc.ListNames(ctx, limit, cursor)
}

func (lm loggingMiddleware) Search(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Search(%v, %v): %s", pkgName, q, cursor, err))
	}()
	return lm.svc.Search(ctx, q, cursor)
}
