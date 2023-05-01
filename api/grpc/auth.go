package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const keyGroupId = "x-awakari-group-id"
const keyUserId = "x-awakari-user-id"

func getAuthInfo(ctx context.Context) (groupId, userId string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Error(codes.Unauthenticated, "missing request metadata")
	}
	if err == nil {
		groupId = getMetadataValue(md, keyGroupId)
		if groupId == "" {
			err = status.Error(codes.Unauthenticated, fmt.Sprintf("missing value for %s in request metadata", keyGroupId))
		}
	}
	if err == nil {
		userId = getMetadataValue(md, keyUserId)
		if userId == "" {
			err = status.Error(codes.Unauthenticated, fmt.Sprintf("missing value for %s in request metadata", keyUserId))
		}
	}
	return
}

func getMetadataValue(md metadata.MD, k string) (v string) {
	var vals []string
	if vals = md.Get(k); len(vals) > 0 {
		v = vals[0]
	}
	return
}
