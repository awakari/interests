package matchers

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

func (c clientMock) Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (resp *MatcherData, err error) {
	resp = &MatcherData{
		Key:         in.Key,
		PatternCode: []byte(in.PatternSrc),
	}
	if in.Key == "fail" {
		err = status.Error(codes.Internal, "")
	} else if in.PatternSrc == "invalid" {
		err = status.Error(codes.InvalidArgument, "")
	} else if in.PatternSrc == "locked" {
		err = status.Error(codes.Unavailable, "")
	}
	return
}

func (c clientMock) LockCreate(ctx context.Context, in *LockCreateRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if string(in.PatternCode) == "fail" {
		err = status.Error(codes.Internal, "fail")
	} else if string(in.PatternCode) == "missing" {
		err = status.Error(codes.NotFound, "missing")
	}
	return
}

func (c clientMock) UnlockCreate(ctx context.Context, in *UnlockCreateRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if string(in.PatternCode) == "fail" {
		err = status.Error(codes.Internal, "fail")
	}
	return
}

func (c clientMock) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	if in.Matcher.Key == "missing" {
		err = status.Error(codes.NotFound, "")
	} else if in.Matcher.Key == "fail" {
		err = status.Error(codes.Unknown, "")
	}
	return
}
