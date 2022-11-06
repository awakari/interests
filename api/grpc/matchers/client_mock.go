package matchers

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	clientMock struct {
	}
)

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (c clientMock) Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (resp *CreateResponse, err error) {
	if in.Key == "fail" {
		err = status.Error(codes.Internal, "")
	} else if in.PatternSrc == "invalid" {
		err = status.Error(codes.InvalidArgument, "")
	} else if in.PatternSrc == "locked" {
		err = status.Error(codes.Unavailable, "")
	} else {
		resp = &CreateResponse{
			Matcher: &MatcherData{
				Key:         in.Key,
				PatternCode: []byte(in.PatternSrc),
			},
		}
	}
	return
}

func (c clientMock) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (resp *DeleteResponse, err error) {
	if in.Matcher.Key == "missing" {
		err = status.Error(codes.NotFound, "")
	} else if in.Matcher.Key == "fail" {
		err = status.Error(codes.Unknown, "")
	}
	resp = &DeleteResponse{}
	return
}

func (c clientMock) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (resp *SearchResponse, err error) {
	if in.Query.Key == "fail" {
		err = status.Error(codes.Internal, "")
	} else {
		resp = &SearchResponse{
			Page: [][]byte{
				[]byte("abc"),
				[]byte("def"),
			},
		}
	}
	return
}
