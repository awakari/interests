package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/util"
	"github.com/google/uuid"
)

type storageMock struct {
}

func NewStorageMock(storage map[string]subscription.Data) Storage {
	return storageMock{}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error) {
	descr := sd.Description
	if descr == "fail" {
		err = ErrInternal
	}
	if err == nil {
		id = uuid.NewString()
	}
	return
}

func (s storageMock) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		sd = subscription.Data{
			Description: "description",
			Condition: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicAnd,
				[]condition.Condition{
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
						"pattern0", false,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(true), "", "key1"),
						"pattern1", false,
					),
				},
			),
		}
	}
	return
}

func (s storageMock) Update(ctx context.Context, id, groupId, userId string, sd subscription.Data) (err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	}
	return
}

func (s storageMock) Delete(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		sd = subscription.Data{
			Description: "description",
			Condition: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicAnd,
				[]condition.Condition{
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
						"pattern0", false,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(true), "", "key1"),
						"pattern1", false,
					),
				},
			),
		}
	}
	return
}

func (s storageMock) SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error) {
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

func (s storageMock) SearchByCondition(ctx context.Context, condId string, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	for i := 0; i < 10_000; i++ {
		cm := subscription.ConditionMatch{
			SubscriptionId: fmt.Sprintf("sub%d", i),
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
				"pattern0", false,
			),
		}
		err = consumeFunc(&cm)
		if err != nil {
			break
		}
	}
	return
}
