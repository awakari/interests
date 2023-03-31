package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/google/uuid"
	"strings"
)

type (
	storageMock struct {
		storage map[string]subscription.Data
	}
)

func NewStorageMock(storage map[string]subscription.Data) Storage {
	return storageMock{
		storage: storage,
	}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, acc string, sd subscription.Data) (id string, err error) {
	id = uuid.NewString()
	s.storage[id+acc] = sd
	return
}

func (s storageMock) Read(ctx context.Context, id, acc string) (sub subscription.Data, err error) {
	var found bool
	sub, found = s.storage[id+acc]
	if !found {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error) {
	sd, found := s.storage[id+acc]
	if found {
		sd.Metadata = md
		s.storage[id+acc] = sd
	} else {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) Delete(ctx context.Context, id, acc string) (sd subscription.Data, err error) {
	var found bool
	sd, found = s.storage[id+acc]
	if found {
		delete(s.storage, id+acc)
	} else {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error) {
	for id, _ := range s.storage {
		if strings.HasSuffix(id, q.Account) && id > cursor {
			ids = append(ids, id[:len(q.Account)])
		}
		if uint32(len(ids)) == q.Limit {
			break
		}
	}
	return
}

func (s storageMock) SearchByKiwi(ctx context.Context, q KiwiQuery, cursor subscription.ConditionMatchKey) (page []subscription.ConditionMatch, err error) {
	for id, sd := range s.storage {
		if containsKiwi(sd.Condition, q.Key, q.Pattern) && id > cursor.Id {
			cm := subscription.ConditionMatch{
				Key: subscription.ConditionMatchKey{
					Id:       id,
					Priority: sd.Metadata.Priority,
				},
				Account:   id,
				Condition: sd.Condition,
			}
			page = append(page, cm)
		}
		if uint32(len(page)) == q.Limit {
			break
		}
	}
	return
}

func containsKiwi(c condition.Condition, k, p string) (contains bool) {
	switch cond := c.(type) {
	case condition.GroupCondition:
		for _, childCond := range cond.GetGroup() {
			contains = containsKiwi(childCond, k, p)
			if contains {
				break
			}
		}
	case condition.KiwiCondition:
		contains = cond.GetKey() == k && cond.GetPattern() == p
	}
	return
}
