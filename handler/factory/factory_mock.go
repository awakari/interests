package factory

import (
	"synapse"
)

type (
	factoryMock struct {
		handler synapse.Handler
	}
)

func NewFactoryMock(handler synapse.Handler) Factory {
	return factoryMock{
		handler: handler,
	}
}

func (f factoryMock) NewHandler(s synapse.Subscription) (synapse.Handler, error) {
	return f.handler, nil
}
