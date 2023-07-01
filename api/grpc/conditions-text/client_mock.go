package conditions_text

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type clientMock struct {
}

func newClientMock() ServiceClient {
	return clientMock{}
}

func (cm clientMock) Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (resp *CreateResponse, err error) {
	switch in.Key {
	case "fail":
		err = status.Error(codes.Internal, "internal failure")
	case "conflict":
		err = status.Error(codes.AlreadyExists, "already exists")
	default:
		resp = &CreateResponse{
			Id:   in.Key,
			Term: in.Term,
		}
	}
	return
}

func (cm clientMock) LockCreate(ctx context.Context, in *LockCreateRequest, opts ...grpc.CallOption) (resp *LockCreateResponse, err error) {
	switch in.Id {
	case "missing":
		err = status.Error(codes.NotFound, "not found")
	default:
		resp = &LockCreateResponse{}
	}
	return
}

func (cm clientMock) UnlockCreate(ctx context.Context, in *UnlockCreateRequest, opts ...grpc.CallOption) (resp *UnlockCreateResponse, err error) {
	switch in.Id {
	case "fail":
		err = status.Error(codes.Internal, "internal failure")
	default:
		resp = &UnlockCreateResponse{}
	}
	return
}

func (cm clientMock) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (resp *DeleteResponse, err error) {
	switch in.Id {
	case "fail":
		err = status.Error(codes.Internal, "internal failure")
	default:
		resp = &DeleteResponse{}
	}
	return
}
