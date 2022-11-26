package storage

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"golang.org/x/exp/slices"
)

type (
	storageMock struct {
		storage map[string]model.Subscription
	}
)

func NewStorageMock(storage map[string]model.Subscription) Storage {
	return storageMock{
		storage: storage,
	}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, sub model.Subscription) (err error) {
	_, found := s.storage[sub.Name]
	if found {
		err = ErrConflict
	} else {
		s.storage[sub.Name] = sub
	}
	return
}

func (s storageMock) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	var found bool
	for k, v := range s.storage {
		if k == name {
			sub = v
			found = true
		}
	}
	if !found {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, name)
	}
	return
}

func (s storageMock) Delete(ctx context.Context, name string) (sub model.Subscription, err error) {
	var found bool
	sub, found = s.storage[name]
	if found {
		delete(s.storage, name)
	} else {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, name)
	}
	return
}

func (s storageMock) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	if cursor == "fail" {
		err = ErrInternal
	} else {
		for k, _ := range s.storage {
			page = append(page, k)
		}
		slices.Sort(page)
	}
	return
}

func (s storageMock) Find(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error) {
	var mg model.MatcherGroup
	for _, sub := range s.storage {
		if q.InExcludes {
			mg = sub.Excludes
		} else {
			mg = sub.Includes
		}
		for _, m := range mg.Matchers {
			if m.Equal(q.Matcher) && sub.Name > cursor {
				page = append(page, sub)
				break
			}
		}
	}
	return
}
