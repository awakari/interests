package service

import (
	"context"
	"github.com/awakari/subscriptions/model"
	"github.com/google/uuid"
)

type (
	serviceMock struct {
	}
)

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Create(ctx context.Context, sd model.SubscriptionData) (id string, err error) {
	descr := sd.Metadata["description"]
	if descr == "fail" {
		err = ErrInternal
	} else if descr == "invalid" {
		err = model.ErrInvalidSubscription
	} else if descr == "conflict" {
		err = ErrConflict
	} else if descr == "busy" {
		err = ErrShouldRetry
	}
	if err == nil {
		id = uuid.NewString()
	}
	return
}

func (sm serviceMock) Read(ctx context.Context, id string) (sd model.SubscriptionData, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		sd = model.SubscriptionData{
			Metadata: map[string]string{
				"description": "description",
			},
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
							"",
							"key0",
						),
						true,
						"pattern0",
					),
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(true),
							"",
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

func (sm serviceMock) Delete(ctx context.Context, id string) (err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	}
	return
}

func (sm serviceMock) SearchByCondition(ctx context.Context, q model.ConditionQuery, cursor string) (page []model.Subscription, err error) {
	if cursor == "" {
		page = []model.Subscription{
			{
				Id: "sub0",
			},
			{
				Id: "sub1",
			},
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}

func (sm serviceMock) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []model.Subscription, err error) {
	if cursor == "" {
		page = []model.Subscription{
			{
				Id: "sub0",
			},
			{
				Id: "sub1",
			},
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}
