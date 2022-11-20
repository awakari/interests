package aggregator

import (
	"context"
	"github.com/meandros-messaging/subscriptions/model"
)

type (
	MatchInGroupMock struct {
		all          bool
		matcherCount uint32
		matchedCount uint32
	}

	serviceMock struct {
		includes map[model.MessageId]map[model.SubscriptionKey]MatchInGroupMock
		excludes map[model.MessageId]map[model.SubscriptionKey]MatchInGroupMock
	}
)

func NewServiceMock(
	includes map[model.MessageId]map[model.SubscriptionKey]MatchInGroupMock,
	excludes map[model.MessageId]map[model.SubscriptionKey]MatchInGroupMock,
) Service {
	return &serviceMock{
		includes: includes,
		excludes: excludes,
	}
}

func (svc *serviceMock) Update(ctx context.Context, m Match) (err error) {
	//
	msgId := m.MessageId
	subKey := m.SubscriptionKey
	//
	incMgBySub, found := svc.includes[msgId]
	if !found {
		incMgBySub = make(map[model.SubscriptionKey]MatchInGroupMock)
		svc.includes[msgId] = incMgBySub
	}
	incMg, found := incMgBySub[subKey]
	if !found {
		incMg = MatchInGroupMock{
			all:          m.Includes.All,
			matcherCount: m.Includes.MatcherCount,
			matchedCount: 0,
		}
	}
	//
	excMgBySub, found := svc.excludes[msgId]
	if !found {
		excMgBySub = make(map[model.SubscriptionKey]MatchInGroupMock)
		svc.excludes[msgId] = excMgBySub
	}
	excMg, found := excMgBySub[subKey]
	if !found {
		excMg = MatchInGroupMock{
			all:          m.Excludes.All,
			matcherCount: m.Excludes.MatcherCount,
			matchedCount: 0,
		}
	}
	//
	if m.InExcludes {
		excMg.matchedCount += 1
		excMgBySub[subKey] = excMg
	} else {
		incMg.matchedCount += 1
		incMgBySub[subKey] = incMg
	}
	//
	return
}
