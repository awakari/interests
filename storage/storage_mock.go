package storage

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/google/uuid"
)

type (
	storageMock struct {
		storage map[string]model.SubscriptionData
	}
)

func NewStorageMock(storage map[string]model.SubscriptionData) Storage {
	return storageMock{
		storage: storage,
	}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, sd model.SubscriptionData) (id string, err error) {
	if sd.Metadata["description"] == "conflict" {
		err = ErrConflict
	} else {
		id = uuid.NewString()
		s.storage[id] = sd
	}
	return
}

func (s storageMock) Read(ctx context.Context, id string) (sub model.SubscriptionData, err error) {
	var found bool
	sub, found = s.storage[id]
	if !found {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) Delete(ctx context.Context, id string) (sd model.SubscriptionData, err error) {
	var found bool
	sd, found = s.storage[id]
	if found {
		delete(s.storage, id)
	} else {
		err = fmt.Errorf("%w by id: %s", ErrNotFound, id)
	}
	return
}

func (s storageMock) SearchByKiwi(ctx context.Context, q KiwiQuery, cursor string) (page []model.Subscription, err error) {
	for id, sd := range s.storage {
		if containsKiwi(sd.Condition, q.Key, q.Pattern) && id > cursor {
			sub := model.Subscription{
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

func containsKiwi(c model.Condition, k, p string) (contains bool) {
	switch cond := c.(type) {
	case model.GroupCondition:
		for _, childCond := range cond.GetGroup() {
			contains = containsKiwi(childCond, k, p)
			if contains {
				break
			}
		}
	case model.KiwiCondition:
		contains = cond.GetKey() == k && cond.GetPattern() == p
	}
	return
}

func (s storageMock) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []model.Subscription, err error) {
	for id, sd := range s.storage {
		if contains(sd.Metadata, q.Metadata) && id > cursor {
			sub := model.Subscription{
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
