package grpc

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
	"io"
	"os"
	"testing"
	"time"
)

const port = 50051

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
		md   []string
		cond *ConditionInput
		err  error
	}{
		"ok1": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
		},
		"ok2": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
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
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},

			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"invalid": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.InvalidArgument, "invalid subscription condition"),
		},
		"busy": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unavailable, "retry the operation"),
		},
		"empty group": {
			md: []string{
				"x-awakari-group-id", "",
				"X-Awakari-User-ID", "user0",
			},
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
		"empty user": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "",
			},
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-user-id in request metadata"),
		},
		"no auth info": {
			cond: &ConditionInput{
				Cond: &ConditionInput_Ktc{
					Ktc: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := metadata.AppendToOutgoingContext(context.TODO(), c.md...)
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
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
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
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
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
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
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
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
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

func TestServiceController_SearchByCondition(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	//
	req := &SearchByConditionRequest{
		Cond: &SearchByConditionRequest_Kcq{
			Kcq: &KiwiConditionQuery{},
		},
	}
	//
	cases := map[string]struct {
		timeout  time.Duration
		minCount int
		maxCount int
		err      error
	}{
		"10 milliseconds is not enough to read 10_000 items": {
			timeout:  10 * time.Millisecond,
			minCount: 1,
			maxCount: 10_000,
			err:      status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
		},
		"10 seconds is more than enough to read 10_000 items": {
			timeout:  10 * time.Second,
			minCount: 10_000,
			maxCount: 10_000,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.Nil(t, err)
			defer conn.Close()
			client := NewServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.TODO(), c.timeout)
			defer cancel()
			stream, err := client.SearchByCondition(ctx, req)
			require.Nil(t, err)
			count := 0
			for {
				_, err = stream.Recv()
				if err == io.EOF {
					err = nil
					break
				}
				if err != nil {
					break
				}
				count++
			}
			assert.ErrorIs(t, err, c.err)
			assert.True(t, count >= c.minCount, count)
			assert.True(t, count <= c.maxCount, count)
		})
	}
}
