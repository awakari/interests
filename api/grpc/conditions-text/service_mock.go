package conditions_text

import (
	"context"
)

type serviceMock struct {
}

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Create(ctx context.Context, k, v string, exact bool) (id, out string, err error) {
	switch k {
	case "fail":
		err = ErrInternal
	case "conflict":
		err = ErrConflict
	case "fail_lock":
		id, out = "fail_lock", v
	default:
		id, out = "cond0", v
	}
	return
}

func (sm serviceMock) LockCreate(ctx context.Context, id string) (err error) {
	switch id {
	case "fail":
		err = ErrInternal
	case "fail_lock":
		err = ErrInternal
	case "missing":
		err = ErrNotFound
	}
	return
}

func (sm serviceMock) UnlockCreate(ctx context.Context, id string) (err error) {
	switch id {
	case "fail":
		err = ErrInternal
	}
	return
}

func (sm serviceMock) Delete(ctx context.Context, id string) (err error) {
	switch id {
	case "fail":
		err = ErrInternal
	}
	return
}
