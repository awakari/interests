package grpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const keyApiKey = "x-api-key"
const keyEndpointApiUserInfo = "x-endpoint-api-userinfo"
const keyAccount = "acc"

type authMiddleware struct {
	srv ServiceServer
}

func NewAuthMiddleware(srv ServiceServer) ServiceServer {
	return authMiddleware{
		srv: srv,
	}
}

func (am authMiddleware) Create(ctx context.Context, request *CreateRequest) (resp *CreateResponse, err error) {
	var authCtx context.Context
	authCtx, err = accountFromContext(ctx)
	if err != nil {
		resp = &CreateResponse{}
	} else {
		resp, err = am.srv.Create(authCtx, request)
	}
	return
}

func (am authMiddleware) Read(ctx context.Context, request *ReadRequest) (resp *ReadResponse, err error) {
	var authCtx context.Context
	authCtx, err = accountFromContext(ctx)
	if err != nil {
		resp = &ReadResponse{}
	} else {
		resp, err = am.srv.Read(authCtx, request)
	}
	return
}

func (am authMiddleware) UpdateMetadata(ctx context.Context, request *UpdateMetadataRequest) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	var authCtx context.Context
	authCtx, err = accountFromContext(ctx)
	if err != nil {
		resp = &emptypb.Empty{}
	} else {
		resp, err = am.srv.UpdateMetadata(authCtx, request)
	}
	return
}

func (am authMiddleware) Delete(ctx context.Context, request *DeleteRequest) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	var authCtx context.Context
	authCtx, err = accountFromContext(ctx)
	if err != nil {
		resp = &emptypb.Empty{}
	} else {
		resp, err = am.srv.Delete(authCtx, request)
	}
	return
}

func (am authMiddleware) SearchOwn(ctx context.Context, request *SearchOwnRequest) (resp *SearchOwnResponse, err error) {
	var authCtx context.Context
	authCtx, err = accountFromContext(ctx)
	if err != nil {
		resp = &SearchOwnResponse{}
	} else {
		resp, err = am.srv.SearchOwn(authCtx, request)
	}
	return
}

func (am authMiddleware) SearchByCondition(ctx context.Context, request *SearchByConditionRequest) (*SearchByConditionResponse, error) {
	return am.srv.SearchByCondition(ctx, request)
}

func accountFromContext(src context.Context) (dst context.Context, err error) {
	md, ok := metadata.FromIncomingContext(src)
	if !ok {
		err = status.Error(codes.Unauthenticated, "missing request metadata")
	} else {
		dst, err = accountFromMetadata(src, md)
	}
	return
}

func accountFromMetadata(src context.Context, md metadata.MD) (dst context.Context, err error) {
	var vals []string
	if vals = md.Get(keyApiKey); len(vals) > 0 && vals[0] != "" {
		dst = context.WithValue(src, keyAccount, vals[0])
	} else if vals = md.Get(keyEndpointApiUserInfo); len(vals) > 0 && vals[0] != "" {
		var acc string
		acc, err = accountFromUserToken(vals[0])
		if err == nil {
			dst = context.WithValue(src, keyAccount, acc)
		}
	} else {
		err = status.Error(
			codes.Unauthenticated,
			fmt.Sprintf(
				"missing request metadata, neither \"%s\" nor \"%s\" set", keyApiKey, keyEndpointApiUserInfo,
			),
		)
	}
	return
}

func accountFromUserToken(tokenBase64 string) (id string, err error) {
	var tokenJsonBytes []byte
	tokenJsonBytes, err = base64.URLEncoding.DecodeString(tokenBase64)
	if err != nil {
		err = status.Error(
			codes.Unauthenticated,
			fmt.Sprintf("invalid user token, failed to decode as Base64 encoded string: %s", err),
		)
	} else {
		var token map[string]any
		err = json.Unmarshal(tokenJsonBytes, &token)
		if err != nil {
			err = status.Error(
				codes.Unauthenticated,
				fmt.Sprintf("invalid user token, failed to parse as JSON the decoded value: %s", err),
			)
		} else {
			id, err = emailFromParsedToken(token)
		}
	}
	return
}

// https://cloud.google.com/endpoints/docs/openapi/migrate-to-esp-v2#handle-jwt
func emailFromParsedToken(token map[string]any) (email string, err error) {
	emailRaw, ok := token["email"]
	if !ok {
		err = status.Error(
			codes.Unauthenticated,
			fmt.Sprintf("invalid user token, missing \"email\" claim"),
		)
	} else {
		email, ok = emailRaw.(string)
		if !ok {
			err = status.Error(
				codes.Unauthenticated,
				fmt.Sprintf("invalid user token, \"email\" claim value type: %T", emailRaw),
			)
		}
	}
	return
}
