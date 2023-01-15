package kiwi

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/model"
)

type (
	serviceMock struct {
	}
)

func NewServiceMock() Service {
	return serviceMock{}
}

func (svc serviceMock) Create(ctx context.Context, k string, patternSrc string) (m model.MatcherData, err error) {
	if k == "fail" {
		err = errors.New("")
	} else if patternSrc == "locked" {
		err = ErrShouldRetry
	} else {
		m = model.MatcherData{
			Key: k,
			Pattern: model.Pattern{
				Code: []byte(patternSrc),
				Src:  patternSrc,
			},
		}
	}
	return
}

func (svc serviceMock) LockCreate(ctx context.Context, patternCode model.PatternCode) (err error) {
	if string(patternCode) == "fail" {
		err = ErrInternal
	}
	return
}

func (svc serviceMock) UnlockCreate(ctx context.Context, patternCode model.PatternCode) (err error) {
	if string(patternCode) == "fail" {
		err = ErrInternal
	}
	return
}

func (svc serviceMock) Delete(ctx context.Context, m model.MatcherData) (err error) {
	if m.Key == "fail" {
		return errors.New("unexpected")
	} else if m.Key == "missing" || string(m.Pattern.Code) == "missing" {
		err = ErrNotFound
	}
	return
}
