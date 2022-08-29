package patterns

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
)

type (
	loggingMiddleware struct {
		log     *logrus.Entry
		svc     Service
		pkgName string
	}
)

func NewLoggingMiddleware(log *logrus.Entry, svc Service) Service {
	return loggingMiddleware{
		log:     log,
		svc:     svc,
		pkgName: reflect.TypeOf(service{}).PkgPath(),
	}
}

func (lm loggingMiddleware) Create(ctx context.Context, src string) (id Id, err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.Create(patternSrc=%s): %s, %s", lm.pkgName, src, id, err))
	}()
	return lm.svc.Create(ctx, src)
}

func (lm loggingMiddleware) Read(ctx context.Context, id Id) (src string, err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.Read(id=%s): %s, %s", lm.pkgName, id, src, err))
	}()
	return lm.svc.Read(ctx, id)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id Id) (err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.Delete(id=%s): %s", lm.pkgName, id, err))
	}()
	return lm.svc.Delete(ctx, id)
}

func (lm loggingMiddleware) SearchMatches(ctx context.Context, input string, limit uint32, cursor Id) (page []Id, err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.SearchMatches(input=%s, limit=%d, cursor=%s): %v, %s", lm.pkgName, input, limit, cursor, page, err))
	}()
	return lm.svc.SearchMatches(ctx, input, limit, cursor)
}
