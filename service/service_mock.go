package service

import (
	"context"
	"github.com/meandros-messaging/subscriptions/model"
)

type (
	serviceMock struct {
	}
)

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Create(ctx context.Context, name string, req CreateRequest) (err error) {
	if name == "fail" {
		err = ErrInternal
	} else if name == "invalid" {
		err = model.ErrInvalidSubscription
	} else if name == "conflict" {
		err = ErrConflict
	} else if name == "busy" {
		err = ErrShouldRetry
	}
	return
}

func (sm serviceMock) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	if name == "fail" {
		err = ErrInternal
	} else if name == "missing" {
		err = ErrNotFound
	} else {
		sub = model.Subscription{
			Name:        name,
			Description: "description",
			Routes: []string{
				"destination",
			},
			Includes: model.MatcherGroup{
				All: true,
				Matchers: []model.Matcher{
					{
						Partial: true,
						MatcherData: model.MatcherData{
							Key: "key0",
							Pattern: model.Pattern{
								Code: []byte("pattern0"),
								Src:  "pattern0",
							},
						},
					},
				},
			},
			Excludes: model.MatcherGroup{
				Matchers: []model.Matcher{
					{
						MatcherData: model.MatcherData{
							Key: "key1",
							Pattern: model.Pattern{
								Code: []byte("pattern1"),
								Src:  "pattern1",
							},
						},
					},
				},
			},
		}
	}
	return
}

func (sm serviceMock) Delete(ctx context.Context, name string) (err error) {
	if name == "fail" {
		err = ErrInternal
	} else if name == "missing" {
		err = ErrNotFound
	}
	return
}

func (sm serviceMock) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	if cursor == "" {
		page = []string{
			"sub0",
			"sub1",
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}

func (sm serviceMock) Search(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error) {
	if cursor == "" {
		page = []model.Subscription{
			{
				Name: "sub0",
			},
			{
				Name: "sub1",
			},
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}
