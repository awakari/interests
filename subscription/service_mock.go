package subscription

import (
	"fmt"
	"math/rand"
	"sort"
	"synapse/util"
)

type (
	serviceMock struct {
		storage map[Id]Data
	}
)

func NewServiceMock(storage map[Id]Data) Service {
	return serviceMock{
		storage: storage,
	}
}

func (svc serviceMock) Create(data Data) (Id, error) {
	id := Id(fmt.Sprintf("%d", rand.Uint64()))
	if _, exists := svc.storage[id]; exists {
		return id, ErrConflict
	}
	svc.storage[id] = data
	return id, nil
}

func (svc serviceMock) Read(id Id) (*Subscription, error) {
	d, exists := svc.storage[id]
	if !exists {
		return nil, ErrNotFound
	}
	return &Subscription{
		Id:   id,
		Data: d,
	}, nil
}

func (svc serviceMock) Delete(id Id) error {
	_, exists := svc.storage[id]
	if !exists {
		return ErrNotFound
	}
	delete(svc.storage, id)
	return nil
}

func (svc serviceMock) List(query SubscriptionsQuery) (SubscriptionsPage, error) {
	ids := make([]string, 0, len(svc.storage))
	for id := range svc.storage {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	cursorPassed := query.CursorRef == nil
	items := make([]Subscription, 0)
	count := uint(0)
	complete := true
	var last *SubscriptionsPageCursor = nil
	for _, id := range ids {
		if !cursorPassed {
			if SubscriptionsPageCursor(id) == *query.CursorRef {
				cursorPassed = true
			}
			continue
		}
		sid := Id(id)
		d := svc.storage[sid]
		if query.TopicIdRef != nil {
			matches := false
			for _, tid := range d.TopicIds {
				if *query.TopicIdRef == tid {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}
		if count == query.Limit {
			complete = false
			break
		}
		item := Subscription{
			Id:   sid,
			Data: svc.storage[sid],
		}
		items = append(items, item)
		count++
		cursor := SubscriptionsPageCursor(id)
		last = &cursor
	}
	return SubscriptionsPage{
		Page: util.Page[Subscription, SubscriptionsPageCursor]{
			Items:     items,
			Complete:  complete,
			CursorRef: last,
		},
	}, nil
}
