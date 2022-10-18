package patterns

import (
	"context"
	"github.com/meandros-messaging/subscriptions/model"
)

type (
	serviceMock struct {
		storage map[string]model.Pattern
	}
)

func NewServiceMock(storage map[string]model.Pattern) Service {
	return &serviceMock{
		storage: storage,
	}
}

func (svc serviceMock) Create(ctx context.Context, src string) (p model.Pattern, _ error) {
	p = model.Pattern{Src: src}
	svc.storage[src] = p
	return
}

func (svc serviceMock) Read(ctx context.Context, code model.PatternCode) (p model.Pattern, err error) {
	src := string(code)
	var found bool
	if p, found = svc.storage[src]; !found {
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) Delete(ctx context.Context, code model.PatternCode) (err error) {
	src := string(code)
	if _, found := svc.storage[src]; found {
		delete(svc.storage, src)
	} else {
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) SearchMatchesBulk(ctx context.Context, md model.Metadata, limit uint32, cursor *BulkCursor) (page map[string][]model.PatternCode, err error) {
	page = make(map[string][]model.PatternCode, len(md))
	for k, v := range md {
		var codes []model.PatternCode
		if p, found := svc.storage[v]; found {
			codes = append(codes, p.Code)
		}
		page[k] = codes
	}
	return
}
