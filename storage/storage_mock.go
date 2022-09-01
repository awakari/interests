package storage

import (
	"context"
)

type (
	storageMock struct {
		storage map[string]map[string][]string
	}
)

func NewStorageMock(storage map[string]map[string][]string) Storage {
	return storageMock{
		storage: storage,
	}
}

func (s storageMock) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s storageMock) Create(ctx context.Context, sub Subscription) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s storageMock) Read(ctx context.Context, id string) (Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (s storageMock) Update(ctx context.Context, id string, sub Subscription) error {
	//TODO implement me
	panic("implement me")
}

func (s storageMock) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (s storageMock) List(ctx context.Context, limit uint32, cursor *string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (s storageMock) Resolve(ctx context.Context, limit uint32, cursor *string, key string, patternIds []PatternCode) (names []string, err error) {
	namesByPatternCode := s.storage[key]
	count := uint32(0)
	for _, patternId := range patternIds {
		if count >= limit {
			break
		}
		for _, name := range namesByPatternCode[string(patternId)] {
			if count >= limit {
				break
			}
			names = append(names, name)
			count++
		}
	}
	return
}
