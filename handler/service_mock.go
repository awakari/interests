package handler

import (
	"synapse"
	"synapse/handler/factory"
)

type (
	serviceMock struct {
		loadFunc func(s synapse.Subscription) (synapse.Handler, error)
		storage  map[synapse.SubscriptionId]synapse.Handler
	}
)

func NewServiceMock(factorySvc factory.Service, storage map[synapse.SubscriptionId]synapse.Handler) Service {
	loadFunc := func(s synapse.Subscription) (synapse.Handler, error) {
		f, err := factorySvc.Get(s.HandlerFactoryName)
		if err != nil {
			return nil, err
		}
		return f.NewHandler(s)
	}
	return serviceMock{
		loadFunc: loadFunc,
		storage:  storage,
	}
}

func (svc serviceMock) Resolve(ss []synapse.Subscription) ([]synapse.Handler, error) {
	hs := make([]synapse.Handler, 0, len(ss))
	for _, s := range ss {
		h, present := svc.storage[s.Id]
		if !present {
			h, _ = svc.loadFunc(s)
			svc.storage[s.Id] = h
		}
		hs = append(hs, h)
	}
	return hs, nil
}
