package kiwiTree

import (
	"context"
	"errors"
)

type (
	serviceMock struct {
	}
)

func NewServiceMock() Service {
	return serviceMock{}
}

func (svc serviceMock) Create(ctx context.Context, k string, pattern string) (err error) {
	if k == "fail" {
		err = errors.New("")
	} else if pattern == "locked" {
		err = ErrShouldRetry
	}
	return
}

func (svc serviceMock) LockCreate(ctx context.Context, k string, pattern string) (err error) {
	if pattern == "fail" {
		err = ErrInternal
	}
	return
}

func (svc serviceMock) UnlockCreate(ctx context.Context, k string, pattern string) (err error) {
	if pattern == "fail" {
		err = ErrInternal
	}
	return
}

func (svc serviceMock) Delete(ctx context.Context, k string, pattern string) (err error) {
	if k == "fail" {
		return errors.New("unexpected")
	} else if k == "missing" || pattern == "missing" {
		err = ErrNotFound
	}
	return
}
