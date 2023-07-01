package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/util"
	"github.com/google/uuid"
	"strings"
)

type storageMock struct {
	storage map[string]subscription.Data
}

func NewStorageMock(storage map[string]subscription.Data) Storage {
	return storageMock{
		storage: storage,
	}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error) {
	id = uuid.NewString()
	s.storage[id+groupId+userId] = sd
	return
}

func (s storageMock) Read(ctx context.Context, id, groupId, userId string) (sub subscription.Data, err error) {
	var found bool
	sub, found = s.storage[id+groupId+userId]
	if !found {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) Update(ctx context.Context, id, groupId, userId string, sd subscription.Data) (err error) {
	sd, found := s.storage[id+groupId+userId]
	if found {
		s.storage[id+groupId+userId] = sd
	} else {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) Delete(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	var found bool
	sd, found = s.storage[id+groupId+userId]
	if found {
		delete(s.storage, id+groupId+userId)
	} else {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error) {
	for id, _ := range s.storage {
		if strings.HasSuffix(id, q.UserId) && id > cursor {
			ids = append(ids, id[:len(q.UserId)])
		}
		if uint32(len(ids)) == q.Limit {
			break
		}
	}
	return
}

func (s storageMock) SearchByCondition(ctx context.Context, condId string, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	for i := 0; i < 10_000; i++ {
		cm := subscription.ConditionMatch{
			SubscriptionId: fmt.Sprintf("sub%d", i),
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
				"pattern0",
				false,
			),
		}
		err = consumeFunc(&cm)
		if err != nil {
			break
		}
	}
	return
}
