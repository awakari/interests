package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"golang.org/x/exp/slices"
)

type (
	storageMock struct {
		storage map[model.SubscriptionKey]model.Subscription
	}
)

func NewStorageMock(storage map[model.SubscriptionKey]model.Subscription) Storage {
	return storageMock{
		storage: storage,
	}
}

func (s storageMock) Close() error {
	return nil
}

func (s storageMock) Create(ctx context.Context, sub model.Subscription) (err error) {
	_, found := s.storage[sub.SubscriptionKey]
	if found {
		err = ErrConflict
	} else {
		s.storage[sub.SubscriptionKey] = sub
	}
	return
}

func (s storageMock) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	var found bool
	for k, v := range s.storage {
		if k.Name == name {
			sub = v
			found = true
		}
	}
	if !found {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, name)
	}
	return
}

func (s storageMock) Update(ctx context.Context, sub model.Subscription) (err error) {
	err = s.DeleteVersion(ctx, sub.SubscriptionKey)
	if err == nil {
		err = s.Create(ctx, sub)
	}
	return
}

func (s storageMock) DeleteVersion(ctx context.Context, subKey model.SubscriptionKey) (err error) {
	var found bool
	_, found = s.storage[subKey]
	if found {
		delete(s.storage, subKey)
	} else {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, subKey.Name)
	}
	return
}

func (s storageMock) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	for k, _ := range s.storage {
		page = append(page, k.Name)
	}
	slices.Sort(page)
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
			if matchersEqual(m, q.Matcher) {
				page = append(page, sub)
				break
			}
		}
	}
	return
}

func matchersEqual(m1, m2 model.Matcher) bool {
	return m1.Partial == m2.Partial && m1.Key == m2.Key && bytes.Equal(m1.Pattern.Code, m2.Pattern.Code)
}
