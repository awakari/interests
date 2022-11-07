package aggregator

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
)

type (
	MatchInGroupMock struct {
		all          bool
		matcherCount uint32
		matchedCount uint32
	}

	serviceMock struct {
		includes *map[model.MessageId]*map[model.SubscriptionKey]MatchInGroupMock
		excludes *map[model.MessageId]*map[model.SubscriptionKey]MatchInGroupMock
	}
)

func NewServiceMock(
	includes *map[model.MessageId]*map[model.SubscriptionKey]MatchInGroupMock,
	excludes *map[model.MessageId]*map[model.SubscriptionKey]MatchInGroupMock,
) Service {
	return &serviceMock{
		includes: includes,
		excludes: excludes,
	}
}

func (svc *serviceMock) Update(ctx context.Context, m Match) (err error) {
	msgId := m.MessageId
	var mgEvt MatchInGroup
	var tbl *map[model.MessageId]*map[model.SubscriptionKey]MatchInGroupMock
	if m.Includes != nil {
		if m.Excludes != nil {
			return errors.New("invalid match event")
		}
		mgEvt = *m.Includes
		tbl = svc.includes
	} else if m.Excludes != nil {
		mgEvt = *m.Excludes
		tbl = svc.excludes
	} else {
		return errors.New("invalid match event")
	}
	mgBySub, found := (*tbl)[msgId]
	if !found {
		t := make(map[model.SubscriptionKey]MatchInGroupMock)
		mgBySub = &t
		(*tbl)[msgId] = mgBySub
	}
	subKey := m.SubscriptionKey
	mg, found := (*mgBySub)[subKey]
	if !found {
		mg = MatchInGroupMock{
			all:          mgEvt.All,
			matcherCount: mgEvt.MatcherCount,
			matchedCount: 0,
		}
	}
	mg.matchedCount += 1
	(*mgBySub)[subKey] = mg
	return
}
