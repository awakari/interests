package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/util"
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

func (s storageMock) Search(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error) {
	var mg model.MatcherGroup
	sortedNames := util.SortedKeys(s.storage)
	for _, subName := range sortedNames {
		sub := s.storage[subName]
		if q.InExcludes {
			mg = sub.Excludes
		} else {
			mg = sub.Includes
		}
		for _, m := range mg.Matchers {
			if matchersEqual(m, q.Matcher) && sub.Name > cursor {
				page = append(page, sub)
				break
			}
		}
		if uint32(len(page)) == q.Limit {
			break
		}
	}
	return
}

// do not replace it with Matcher.Equal because Pattern.Source is missing in the matchers response
func matchersEqual(m1, m2 model.Matcher) bool {
	return m1.Partial == m2.Partial && m1.Key == m2.Key && bytes.Equal(m1.Pattern.Code, m2.Pattern.Code)
}
