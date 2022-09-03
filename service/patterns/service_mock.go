package patterns

import (
	"context"
	"github.com/meandros-messaging/subscriptions/storage"
)

type (
	serviceMock struct {
		storage map[string]Pattern
	}
)

func NewServiceMock(storage map[string]Pattern) Service {
	return &serviceMock{
		storage: storage,
	}
}

func (svc serviceMock) Create(ctx context.Context, src string) (PatternCode, error) {
	svc.storage[src] = Pattern{Src: src}
	return PatternCode(src), nil
}

func (svc serviceMock) Read(ctx context.Context, code PatternCode) (p Pattern, err error) {
	src := string(code)
	var found bool
	if p, found = svc.storage[src]; !found {
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) Delete(ctx context.Context, code PatternCode) (err error) {
	src := string(code)
	if _, found := svc.storage[src]; found {
		delete(svc.storage, src)
	} else {
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) SearchMatchesBulk(ctx context.Context, md storage.Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Pattern, err error) {
	page = make(map[string][]Pattern, len(md))
	for k, v := range md {
		var ps []Pattern
		if p, found := svc.storage[v]; found {
			ps = append(ps, p)
		}
		page[k] = ps
	}
	return
}
