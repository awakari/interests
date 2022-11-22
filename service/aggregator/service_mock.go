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
		includes map[model.MessageId]map[string]MatchInGroupMock
		excludes map[model.MessageId]map[string]MatchInGroupMock
	}
)

func NewServiceMock(
	includes map[model.MessageId]map[string]MatchInGroupMock,
	excludes map[model.MessageId]map[string]MatchInGroupMock,
) Service {
	return &serviceMock{
		includes: includes,
		excludes: excludes,
	}
}

func (svc *serviceMock) Update(ctx context.Context, m Match) (err error) {
	//
	msgId := m.MessageId
	subName := m.SubscriptionName
	//
	incMgBySub, found := svc.includes[msgId]
	if !found {
		incMgBySub = make(map[string]MatchInGroupMock)
		svc.includes[msgId] = incMgBySub
	}
	incMg, found := incMgBySub[subName]
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
		excMgBySub = make(map[string]MatchInGroupMock)
		svc.excludes[msgId] = excMgBySub
	}
	excMg, found := excMgBySub[subName]
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
		excMgBySub[subName] = excMg
	} else {
		incMg.matchedCount += 1
		incMgBySub[subName] = incMg
	}
	//
	return
}
