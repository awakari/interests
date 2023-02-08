package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/google/uuid"
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

func (s storageMock) Create(ctx context.Context, sd subscription.Data) (id string, err error) {
	if sd.Metadata["description"] == "conflict" {
		err = ErrConflict
	} else {
		id = uuid.NewString()
		s.storage[id] = sd
	}
	return
}

func (s storageMock) Read(ctx context.Context, id string) (sub subscription.Data, err error) {
	var found bool
	sub, found = s.storage[id]
	if !found {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) Delete(ctx context.Context, id string) (sd subscription.Data, err error) {
	var found bool
	sd, found = s.storage[id]
	if found {
		delete(s.storage, id)
	} else {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) SearchByKiwi(ctx context.Context, q KiwiQuery, cursor string) (page []subscription.ConditionMatch, err error) {
	for id, sd := range s.storage {
		if containsKiwi(sd.Route.Condition, q.Key, q.Pattern) && id > cursor {
			cm := subscription.ConditionMatch{
				Id:    id,
				Route: sd.Route,
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

func (s storageMock) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []subscription.Subscription, err error) {
	for id, sd := range s.storage {
		if contains(sd.Metadata, q.Metadata) && id > cursor {
			sub := subscription.Subscription{
				Id:   id,
				Data: sd,
			}
			page = append(page, sub)
		}
		if uint32(len(page)) == q.Limit {
			break
		}
	}
	return
}

func contains(a, b map[string]string) bool {
	for k, bv := range b {
		av, present := a[k]
		if !present {
			return false
		}
		if av != bv {
			return false
		}
	}
	return true
}
