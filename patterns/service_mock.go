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

func (svc serviceMock) SearchMatches(ctx context.Context, input string, limit uint32, cursor Code) (codes []Code, err error) {
	if svc.storage[input] {
		codes = append(codes, Code(input))
	}
	return
}
