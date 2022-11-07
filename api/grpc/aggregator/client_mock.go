package aggregator

import (
	"context"
	"google.golang.org/grpc"
)

type (
	clientMock struct {
	}
)

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (c clientMock) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (resp *UpdateResponse, err error) {
	//TODO implement me
	resp = &UpdateResponse{}
	return
}
