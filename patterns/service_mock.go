package patterns

import "context"

type (
	serviceMock struct {
		storage map[string]bool
	}
)

func NewServiceMock(storage map[string]bool) Service {
	return &serviceMock{
		storage: storage,
	}
}

func (svc serviceMock) Create(ctx context.Context, src string) (Code, error) {
	svc.storage[src] = true
	return Code(src), nil
}

func (svc serviceMock) Read(ctx context.Context, code Code) (src string, err error) {
	src = string(code)
	if !svc.storage[src] {
		src = ""
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) Delete(ctx context.Context, code Code) (err error) {
	src := string(code)
	if svc.storage[src] {
		delete(svc.storage, src)
	} else {
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) SearchMatchesBulk(ctx context.Context, md Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Code, err error) {
	page = make(map[string][]Code, len(md))
	for k, v := range md {
		var codes []Code
		if svc.storage[v] {
			codes = append(codes, Code(v))
		}
		page[k] = codes
	}
	return
}
