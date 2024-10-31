package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"time"
)

type storageMock struct {
}

func NewStorageMock(storage map[string]subscription.Data) Storage {
	return storageMock{}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, id, groupId, userId string, sd subscription.Data) (err error) {
	switch id {
	case "fail":
		err = ErrInternal
	case "conflict":
		err = ErrConflict
	}
	return
}

func (s storageMock) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, ownerGroupId, ownerUserId string, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		sd = subscription.Data{
			Description: "description",
			Enabled:     true,
			Expires:     time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
			Created:     time.Date(2024, 4, 9, 7, 3, 25, 0, time.UTC),
			Updated:     time.Date(2024, 4, 9, 7, 3, 35, 0, time.UTC),
			Result:      time.Date(2024, 4, 9, 7, 3, 45, 0, time.UTC),
			Public:      true,
			Followers:   42,
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
					condition.NewNumberCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "", "key2"),
						condition.NumOpEq, 42,
					),
				},
			),
		}
		ownerGroupId = groupId
		ownerUserId = userId
	}
	return
}

func (s storageMock) Update(ctx context.Context, id, groupId, userId string, sd subscription.Data) (prev subscription.Data, err error) {
	if id == "fail" {
		err = ErrInternal
	} else if id == "missing" {
		err = ErrNotFound
	} else {
		prev = subscription.Data{
			Description: "description",
			Expires:     time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
			Condition: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicAnd,
				[]condition.Condition{
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "txt_0", "key0"),
						"pattern0", false,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(true), "txt_1", "key1"),
						"pattern1", false,
					),
				},
			),
		}
	}
	return
}

func (s storageMock) UpdateFollowers(ctx context.Context, id string, count int64) (err error) {
	switch id {
	case "missing":
		err = ErrNotFound
	case "fail":
		err = ErrInternal
	}
	return
}

func (s storageMock) UpdateResultTime(ctx context.Context, id string, last time.Time) (err error) {
	switch id {
	case "missing":
		err = ErrNotFound
	case "fail":
		err = ErrInternal
	}
	return
}

func (s storageMock) SetEnabledBatch(ctx context.Context, ids []string, enabled bool) (n int64, err error) {
	if len(ids) > 0 && ids[0] == "fail" {
		err = ErrInternal
	} else {
		n = int64(len(ids))
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
			Expires:     time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
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

func (s storageMock) Search(ctx context.Context, q subscription.Query, cursor subscription.Cursor) (ids []string, err error) {
	if cursor.Id == "" {
		switch q.Order {
		case subscription.OrderDesc:
			switch q.Sort {
			case subscription.SortFollowers:
				ids = []string{
					"sub0",
					"sub1",
				}
			default:
				ids = []string{
					"sub1",
					"sub0",
				}
			}
		default:
			ids = []string{
				"sub0",
				"sub1",
			}
		}
	} else if cursor.Id == "fail" {
		err = ErrInternal
	}
	return
}

func (s storageMock) SearchByCondition(ctx context.Context, q subscription.QueryByCondition, cursor string) (page []subscription.ConditionMatch, err error) {
	for i := 0; i < int(q.Limit); i++ {
		cm := subscription.ConditionMatch{
			SubscriptionId: fmt.Sprintf("sub%d", i),
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
				"pattern0", false,
			),
		}
		page = append(page, cm)
	}
	return
}

func (s storageMock) Count(ctx context.Context) (count int64, err error) {
	count = 42
	return
}

func (s storageMock) CountUsersUnique(ctx context.Context) (count int64, err error) {
	count = 42
	return
}
