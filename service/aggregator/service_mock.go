package aggregator

import (
	"context"
)

type (
	serviceMock struct {
		sink chan<- Match
	}
)

func NewServiceMock(sink chan<- Match) Service {
	return serviceMock{
		sink: sink,
	}
}

func (svc serviceMock) Enroll(ctx context.Context, m Match) (err error) {
	if m.SubscriptionName == "fail" {
		err = ErrInternal
	} else {
		svc.sink <- m
	}
	return
}
