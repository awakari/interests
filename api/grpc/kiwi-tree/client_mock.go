package kiwiTree

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	clientMock struct {
	}
)

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (c clientMock) Create(ctx context.Context, in *KeyPatternRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if in.Key == "fail" {
		err = status.Error(codes.Internal, "")
	} else if in.Pattern == "invalid" {
		err = status.Error(codes.InvalidArgument, "")
	} else if in.Pattern == "locked" {
		err = status.Error(codes.Unavailable, "")
	}
	return
}

func (c clientMock) LockCreate(ctx context.Context, in *KeyPatternRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if in.Pattern == "fail" {
		err = status.Error(codes.Internal, "fail")
	} else if in.Pattern == "missing" {
		err = status.Error(codes.NotFound, "missing")
	}
	return
}

func (c clientMock) UnlockCreate(ctx context.Context, in *KeyPatternRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if in.Pattern == "fail" {
		err = status.Error(codes.Internal, "fail")
	}
	return
}

func (c clientMock) Delete(ctx context.Context, in *KeyPatternRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if in.Key == "missing" {
		err = status.Error(codes.NotFound, "")
	} else if in.Key == "fail" {
		err = status.Error(codes.Unknown, "")
	}
	return
}
