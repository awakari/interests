package factory

import (
	"synapse/handler"
	"synapse/subscription"
)

type (
	factoryMock struct {
		handler handler.Handler
	}
)

func NewFactoryMock(handler handler.Handler) Factory {
	return factoryMock{
		handler: handler,
	}
}

func (f factoryMock) NewHandler(s subscription.Subscription) (handler.Handler, error) {
	return f.handler, nil
}
