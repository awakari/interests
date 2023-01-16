package service

import (
	"context"
	"github.com/awakari/subscriptions/model"
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
			Condition: model.NewGroupCondition(
				model.NewCondition(false),
				model.GroupLogicAnd,
				[]model.Condition{
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"key0",
						),
						true,
						"pattern0",
					),
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(true),
							"key1",
						),
						false,
						"pattern1",
					),
				},
			),
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

func (sm serviceMock) SearchByKiwi(ctx context.Context, q model.KiwiQuery, cursor string) (page []model.Subscription, err error) {
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
