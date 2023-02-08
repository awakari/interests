package service

import (
	"context"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/google/uuid"
)

type (
	serviceMock struct {
	}
)

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Create(ctx context.Context, sd subscription.Data) (id string, err error) {
	descr := sd.Metadata["description"]
	if descr == "fail" {
		err = ErrInternal
	} else if descr == "invalid" {
		err = subscription.ErrInvalidSubscriptionRoute
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

func (sm serviceMock) Read(ctx context.Context, id string) (sd subscription.Data, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		sd = subscription.Data{
			Metadata: map[string]string{
				"description": "description",
			},
			Route: subscription.Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(
							condition.NewKeyCondition(
								condition.NewCondition(false),
								"",
								"key0",
							),
							true,
							"pattern0",
						),
						condition.NewKiwiCondition(
							condition.NewKeyCondition(
								condition.NewCondition(true),
								"",
								"key1",
							),
							false,
							"pattern1",
						),
					},
				),
			},
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

func (sm serviceMock) SearchByCondition(ctx context.Context, q condition.Query, cursor string) (page []subscription.ConditionMatch, err error) {
	if cursor == "" {
		page = []subscription.ConditionMatch{
			{
				Id:          "sub0",
				ConditionId: "cond0",
			},
			{
				Id:          "sub1",
				ConditionId: "cond0",
			},
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}

func (sm serviceMock) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []subscription.Subscription, err error) {
	if cursor == "" {
		page = []subscription.Subscription{
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
