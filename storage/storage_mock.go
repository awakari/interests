package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/util"
	"strings"
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

func (s storageMock) Create(ctx context.Context, sub model.Subscription) error {
	s.storage[sub.Name] = sub
	return nil
}

func (s storageMock) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	var found bool
	sub, found = s.storage[name]
	if !found {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, name)
	}
	return
}

func (s storageMock) Update(ctx context.Context, sub model.Subscription) (err error) {
	var subOld model.Subscription
	var found bool
	name := sub.Name
	subOld, found = s.storage[name]
	if found && subOld.Version == sub.Version {
		s.storage[name] = sub
	} else {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, name)
	}
	return
}

func (s storageMock) Delete(ctx context.Context, name string) (err error) {
	var found bool
	_, found = s.storage[name]
	if found {
		delete(s.storage, name)
	} else {
		err = fmt.Errorf("%w by name: %s", ErrNotFound, name)
	}
	return
}

func (s storageMock) List(ctx context.Context, limit uint32, cursor *string) (page []string, err error) {
	sortedNames := util.SortedKeys(s.storage)
	for _, name := range sortedNames {
		if uint32(len(page)) >= limit {
			break
		}
		if cursor == nil || strings.Compare(*cursor, name) < 0 {
			page = append(page, name)
		}
	}
	return
}

func (s storageMock) FindCandidates(ctx context.Context, limit uint32, cursor *string, key string, patternCode model.PatternCode) (page []model.Subscription, err error) {
	sortedNames := util.SortedKeys(s.storage)
	for _, name := range sortedNames {
		if uint32(len(page)) >= limit {
			break
		}
		if cursor == nil || strings.Compare(*cursor, name) < 0 {
			sub := s.storage[name]
			for _, matcher := range sub.Includes.Matchers {
				if matches(matcher, key, patternCode) {
					page = append(page, sub)
				}
			}
		}
	}
	return
}

func matches(matcher model.Matcher, key string, patternCode model.PatternCode) bool {
	return key == matcher.Key && bytes.Equal(patternCode, matcher.Pattern.Code)
}
