package public

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/api/grpc/common"
	"github.com/awakari/subscriptions/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"os"
	"testing"
)

const port = 8080

var log = slog.Default()

func TestMain(m *testing.M) {
	svc := service.NewServiceMock()
	svc = service.NewLoggingMiddleware(svc, log)
	go func() {
		err := Serve(svc, port)
		if err != nil {
			log.Error("", err)
		}
	}()
	code := m.Run()
	os.Exit(code)
}

func TestServiceController_Create(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		authKey string
		authVal string
		cond    *ConditionInput
		err     error
	}{
		"ok1": {
			authKey: "X-APi-Key",
			authVal: "yohoho",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
		},
		"ok2": {
			authKey: "X-Endpoint-Api-Userinfo",
			authVal: "eyAiZW1haWwiOiAieW9ob2hvQGVtYWlsLmNvbSIgfQ",
			cond: &ConditionInput{
				Not: false,
				Cond: &ConditionInput_Gc{
					Gc: &GroupConditionInput{
						Logic: common.GroupLogic_And,
						Group: []*ConditionInput{
							{
								Not: true,
								Cond: &ConditionInput_Ktc{
									Ktc: &KiwiTreeConditionInput{
										Key:     "key0",
										Pattern: "pattern0",
										Partial: true,
									},
								},
							},
							{
								Not: false,
								Cond: &ConditionInput_Ktc{
									Ktc: &KiwiTreeConditionInput{
										Key:     "key1",
										Pattern: "pattern1",
										Partial: false,
									},
								},
							},
						},
					},
				},
			},
		},
		"fail": {
			authKey: "X-APi-Key",
			authVal: "yohoho",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"invalid": {
			authKey: "X-APi-Key",
			authVal: "yohoho",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.InvalidArgument, "invalid subscription condition"),
		},
		"busy": {
			authKey: "X-APi-Key",
			authVal: "yohoho",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unavailable, "retry the operation"),
		},
		"empty api key token": {
			authKey: "X-Api-Key",
			authVal: "",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
		"empty user token": {
			authKey: "X-Endpoint-Api-UserInfo",
			authVal: "",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
		"no auth tokens": {
			authKey: "foo",
			authVal: "",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
		"invalid user token": {
			authKey: "X-Endpoint-Api-UserInfo",
			authVal: "eyAiZW1haWwiOiB7ICJmb28iOiAiYmFyIiB9IH0",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "invalid user token, \"email\" claim value type: map[string]interface {}"),
		},
		"invalid user token base64": {
			authKey: "X-Endpoint-Api-UserInfo",
			authVal: "Z",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "invalid user token, failed to decode as Base64 encoded string: illegal base64 data at input byte 0"),
		},
		"invalid user token json": {
			authKey: "X-Endpoint-Api-UserInfo",
			authVal: "bm90IGEganNvbg",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "invalid user token, failed to parse as JSON the decoded value: invalid character 'o' in literal null (expecting 'u')"),
		},
		"invalid user token - no email": {
			authKey: "X-Endpoint-Api-UserInfo",
			authVal: "eyAic3ViIjogMTIzNDUsICJpc3MiOiAiZ29vZ2xlLmNvbSIgfQ",
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "invalid user token, missing \"email\" claim"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := metadata.AppendToOutgoingContext(context.TODO(), c.authKey, c.authVal)
			_, err = client.Create(ctx, &CreateRequest{
				Md: &Metadata{
					Description: k,
				},
				Cond: c.cond,
			})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Read(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth bool
		sub  *ReadResponse
		err  error
	}{
		"ok": {
			auth: true,
			sub: &ReadResponse{
				Md: &Metadata{
					Description: "description",
				},
				Cond: &common.ConditionOutput{
					Not: false,
					Cond: &common.ConditionOutput_Gc{
						Gc: &common.GroupConditionOutput{
							Logic: common.GroupLogic_And,
							Group: []*common.ConditionOutput{
								{
									Not: false,
									Cond: &common.ConditionOutput_Kc{
										Kc: &common.KiwiConditionOutput{
											Key:     "key0",
											Pattern: "pattern0",
											Partial: true,
										},
									},
								},
								{
									Not: true,
									Cond: &common.ConditionOutput_Kc{
										Kc: &common.KiwiConditionOutput{
											Key:     "key1",
											Pattern: "pattern1",
											Partial: false,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			auth: true,
			err:  status.Error(codes.NotFound, "subscription was not found"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "X-Endpoint-Api-UserInfo", "eyAiZW1haWwiOiAieW9ob2hvQGVtYWlsLmNvbSIgfQ")
			}
			sub, err := client.Read(ctx, &ReadRequest{Id: k})
			if c.err == nil {
				require.Nil(t, err)
				assert.Equal(t, c.sub.Md, sub.Md)
				assert.Equal(t, c.sub.Cond.Not, sub.Cond.Not)
				assert.Equal(t, c.sub.Cond.GetGc().Logic, sub.Cond.GetGc().Logic)
				assert.Equal(t, len(c.sub.Cond.GetGc().GetGroup()), len(sub.Cond.GetGc().GetGroup()))
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].Not, sub.Cond.GetGc().GetGroup()[0].Not)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetKc().Key, sub.Cond.GetGc().GetGroup()[0].GetKc().Key)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetKc().Pattern, sub.Cond.GetGc().GetGroup()[0].GetKc().Pattern)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetKc().Partial, sub.Cond.GetGc().GetGroup()[0].GetKc().Partial)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].Not, sub.Cond.GetGc().GetGroup()[1].Not)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].GetKc().Key, sub.Cond.GetGc().GetGroup()[1].GetKc().Key)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].GetKc().Pattern, sub.Cond.GetGc().GetGroup()[1].GetKc().Pattern)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].GetKc().Partial, sub.Cond.GetGc().GetGroup()[1].GetKc().Partial)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_UpdateMetadata(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth bool
		md   Metadata
		err  error
	}{
		"ok1": {
			auth: true,
		},
		"ok2": {
			auth: true,
			md: Metadata{
				Description: "new description",
				Enabled:     true,
			},
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			auth: true,
			err:  status.Error(codes.NotFound, "subscription was not found"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "X-Api-Key", "api-key-value...")
			}
			_, err := client.UpdateMetadata(ctx, &UpdateMetadataRequest{
				Id: k,
				Md: &c.md,
			})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Delete(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth bool
		err  error
	}{
		"ok": {
			auth: true,
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			auth: true,
			err:  status.Error(codes.NotFound, "subscription was not found"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "X-Api-Key", "api-key-value...")
			}
			_, err := client.Delete(ctx, &DeleteRequest{Id: k})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_SearchByAccount(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth bool
		err  error
		ids  []string
	}{
		"": {
			auth: true,
			ids: []string{
				"sub0",
				"sub1",
			},
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal failure"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing request metadata, neither \"x-api-key\" nor \"x-endpoint-api-userinfo\" set"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "X-Api-Key", "api-key-value...")
			}
			resp, err := client.SearchOwn(ctx, &SearchOwnRequest{Cursor: k, Limit: 0})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.ids, resp.Ids)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
