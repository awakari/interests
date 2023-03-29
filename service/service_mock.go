package service

import (
	"context"
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

func (sm serviceMock) Create(ctx context.Context, acc string, sd subscription.Data) (id string, err error) {
	descr := sd.Metadata.Description
	if descr == "fail" {
		err = ErrInternal
	} else if descr == "invalid" {
		err = subscription.ErrInvalidSubscriptionCondition
	} else if descr == "busy" {
		err = ErrShouldRetry
	}
	if err == nil {
		id = uuid.NewString()
	}
	return
}

func (sm serviceMock) Read(ctx context.Context, id, acc string) (sd subscription.Data, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		sd = subscription.Data{
			Metadata: subscription.Metadata{
				Description: "description",
			},
			Condition: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicAnd,
				[]condition.Condition{
					condition.NewKiwiCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
						true,
						"pattern0",
					),
					condition.NewKiwiCondition(
						condition.NewKeyCondition(condition.NewCondition(true), "", "key1"),
						false,
						"pattern1",
					),
				},
			),
		}
	}
	return
}

func (sm serviceMock) UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	}
	return
}

func (sm serviceMock) Delete(ctx context.Context, id, acc string) (err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	}
	return
}

func (sm serviceMock) SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error) {
	if cursor == "" {
		ids = []string{
			"sub0",
			"sub1",
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}

func (sm serviceMock) SearchByCondition(ctx context.Context, q condition.Query, cursor string) (page []subscription.ConditionMatch, err error) {
	if cursor == "" {
		page = []subscription.ConditionMatch{
			{
				Id:      "sub0",
				Account: "acc0",
				Condition: condition.NewKiwiCondition(
					condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
					false,
					"pattern0",
				),
				ConditionId: "cond0",
			},
			{
				Id:      "sub1",
				Account: "acc1",
				Condition: condition.NewKiwiCondition(
					condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
					false,
					"pattern0",
				),
				ConditionId: "cond0",
			},
		}
	} else if cursor == "fail" {
		err = ErrInternal
	}
	return
}
