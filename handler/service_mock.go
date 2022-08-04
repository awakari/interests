package handler

import (
	"synapse/handler/factory"
	"synapse/subscription"
)

type (
	serviceMock struct {
		loadFunc func(s subscription.Subscription) (Handler, error)
		storage  map[subscription.Id]Handler
	}
)

func NewServiceMock(factorySvc factory.Service, storage map[subscription.Id]Handler) Service {
	loadFunc := func(s subscription.Subscription) (Handler, error) {
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

func (svc serviceMock) Resolve(ss []subscription.Subscription) ([]Handler, error) {
	hs := make([]Handler, 0, len(ss))
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
