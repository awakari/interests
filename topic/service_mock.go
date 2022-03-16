package topic

import (
	"sort"
	"strings"
	"synapse"
	"synapse/subscription"
	"synapse/util"
)

type (
	serviceMock struct {
		storage         map[synapse.TopicId]synapse.TopicData
		subscriptionSvc subscription.Service
	}
)

func NewServiceMock(storage map[synapse.TopicId]synapse.TopicData, subscriptionSvc subscription.Service) Service {
	return serviceMock{
		storage:         storage,
		subscriptionSvc: subscriptionSvc,
	}
}

func (svc serviceMock) Create(data synapse.TopicData) (synapse.TopicId, error) {
	id := synapse.TopicId(data.Name)
	if _, exists := svc.storage[id]; exists {
		return id, ErrConflict
	}
	svc.storage[id] = data
	return id, nil
}

func (svc serviceMock) Read(name string) (*synapse.Topic, error) {
	id := synapse.TopicId(name)
	d, exists := svc.storage[id]
	if !exists {
		return nil, ErrNotFound
	}
	return &synapse.Topic{
		Id:        id,
		TopicData: d,
	}, nil
}

func (svc serviceMock) Delete(name string) error {
	id := synapse.TopicId(name)
	_, exists := svc.storage[id]
	if !exists {
		return ErrNotFound
	}
	delete(svc.storage, id)
	return nil
}

func (svc serviceMock) List(query synapse.TopicsQuery) (synapse.TopicsPage, error) {
	ids := make([]string, 0, len(svc.storage))
	for id := range svc.storage {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	cursorPassed := query.CursorRef == nil
	items := make([]synapse.Topic, 0)
	count := uint(0)
	complete := true
	var last *synapse.TopicsPageCursor = nil
	for _, id := range ids {
		if !cursorPassed {
			if synapse.TopicsPageCursor(id) == *query.CursorRef {
				cursorPassed = true
			}
			continue
		}
		if query.NamePrefix != "" && !strings.HasPrefix(id, query.NamePrefix) {
			continue
		}
		if count == query.Limit {
			complete = false
			break
		}
		tid := synapse.TopicId(id)
		item := synapse.Topic{
			Id:        tid,
			TopicData: svc.storage[tid],
		}
		items = append(items, item)
		count++
		cursor := synapse.TopicsPageCursor(id)
		last = &cursor
	}
	return synapse.TopicsPage{
		Page: util.Page[synapse.Topic, synapse.TopicsPageCursor]{
			Items:     items,
			Complete:  complete,
			CursorRef: last,
		},
	}, nil
}
