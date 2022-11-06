package matchers

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/sirupsen/logrus"
	"reflect"
)

type (
	loggingMiddleware struct {
		svc     Service
		log     *logrus.Entry
		pkgName string
	}
)

func NewLoggingMiddleware(svc Service, log *logrus.Entry) Service {
	return loggingMiddleware{
		svc:     svc,
		log:     log,
		pkgName: reflect.TypeOf(loggingMiddleware{}).PkgPath(),
	}
}

func (lm loggingMiddleware) Create(ctx context.Context, k string, patternSrc string) (m model.MatcherData, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Create(%s, %s): %s", lm.pkgName, k, patternSrc, err))
	}()
	return lm.svc.Create(ctx, k, patternSrc)
}

func (lm loggingMiddleware) Delete(ctx context.Context, m model.MatcherData) (err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Delete(%s): %s", lm.pkgName, m, err))
	}()
	return lm.svc.Delete(ctx, m)
}

func (lm loggingMiddleware) Search(ctx context.Context, k, v string, limit uint32, cursor model.PatternCode) (page []model.PatternCode, err error) {
	defer func() {
		lm.log.Debug(fmt.Sprintf("%s.Search(%s, %s, %d, %s): %s", lm.pkgName, k, v, limit, cursor, err))
	}()
	return lm.svc.Search(ctx, k, v, limit, cursor)
}
