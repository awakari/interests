package matchers

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
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

func (svc serviceMock) Delete(ctx context.Context, m model.MatcherData) (err error) {
	if m.Key == "fail" {
		return errors.New("unexpected")
	} else if m.Key == "missing" || string(m.Pattern.Code) == "missing" {
		err = ErrNotFound
	}
	return
}

func (svc serviceMock) Search(ctx context.Context, k, v string, limit uint32, cursor model.PatternCode) (page []model.PatternCode, err error) {
	if k == "fail" {
		err = ErrInternal
	} else {
		page = []model.PatternCode{
			[]byte("abc"),
			[]byte("def"),
		}
	}
	return
}
