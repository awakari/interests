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

func (lm loggingMiddleware) Create(ctx context.Context, src string) (id PatternCode, err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.Create(patternSrc=%s): %s, %s", lm.pkgName, src, id, err))
	}()
	return lm.svc.Create(ctx, src)
}

func (lm loggingMiddleware) Read(ctx context.Context, id PatternCode) (p Pattern, err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.Read(id=%s): %s, %s", lm.pkgName, id, p, err))
	}()
	return lm.svc.Read(ctx, id)
}

func (lm loggingMiddleware) Delete(ctx context.Context, id PatternCode) (err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.Delete(id=%s): %s", lm.pkgName, id, err))
	}()
	return lm.svc.Delete(ctx, id)
}

func (lm loggingMiddleware) SearchMatchesBulk(ctx context.Context, md Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Pattern, err error) {
	defer func() {
		lm.log.Info(fmt.Sprintf("%s.SearchMatchesBulk(md=%s, limit=%d, cursor=%s): %s", lm.pkgName, md, limit, cursor, err))
	}()
	return lm.svc.SearchMatchesBulk(ctx, md, limit, cursor)
}
