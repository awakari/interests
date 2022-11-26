package aggregator

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	clientMock struct {
	}
)

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (c clientMock) Enroll(ctx context.Context, in *EnrollRequest, opts ...grpc.CallOption) (empty *emptypb.Empty, err error) {
	//TODO implement me
	return
}
