package subscription

import (
	"fmt"
	"math/rand"
	"sort"
	"synapse"
	"synapse/util"
)

type (
	serviceMock struct {
		storage map[synapse.SubscriptionId]synapse.SubscriptionData
	}
)

func NewServiceMock(storage map[synapse.SubscriptionId]synapse.SubscriptionData) Service {
	return serviceMock{
		storage: storage,
	}
}

func (svc serviceMock) Create(data synapse.SubscriptionData) (synapse.SubscriptionId, error) {
	id := synapse.SubscriptionId(fmt.Sprintf("%d", rand.Uint64()))
	if _, exists := svc.storage[id]; exists {
		return id, ErrConflict
	}
	svc.storage[id] = data
	return id, nil
}

func (svc serviceMock) Read(id synapse.SubscriptionId) (*synapse.Subscription, error) {
	d, exists := svc.storage[id]
	if !exists {
		return nil, ErrNotFound
	}
	return &synapse.Subscription{
		Id:               id,
		SubscriptionData: d,
	}, nil
}

func (svc serviceMock) Delete(id synapse.SubscriptionId) error {
	_, exists := svc.storage[id]
	if !exists {
		return ErrNotFound
	}
	delete(svc.storage, id)
	return nil
}

func (svc serviceMock) List(query synapse.SubscriptionsQuery) (synapse.SubscriptionsPage, error) {
	ids := make([]string, 0, len(svc.storage))
	for id := range svc.storage {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	cursorPassed := query.CursorRef == nil
	items := make([]synapse.Subscription, 0)
	count := uint(0)
	complete := true
	var last *synapse.SubscriptionsPageCursor = nil
	for _, id := range ids {
		if !cursorPassed {
			if synapse.SubscriptionsPageCursor(id) == *query.CursorRef {
				cursorPassed = true
			}
			continue
		}
		sid := synapse.SubscriptionId(id)
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
		item := synapse.Subscription{
			Id:               sid,
			SubscriptionData: svc.storage[sid],
		}
		items = append(items, item)
		count++
		cursor := synapse.SubscriptionsPageCursor(id)
		last = &cursor
	}
	return synapse.SubscriptionsPage{
		Page: util.Page[synapse.Subscription, synapse.SubscriptionsPageCursor]{
			Items:     items,
			Complete:  complete,
			CursorRef: last,
		},
	}, nil
}
