package grpc

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/api/grpc/common"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
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
	stor := storage.NewStorageMock(make(map[string]subscription.Data))
	stor = storage.NewLoggingMiddleware(stor, log)
	go func() {
		err := Serve(stor, port)
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
		cond *Condition
		err  error
	}{
		"ok1": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
		},
		"ok2": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Not: false,
				Cond: &Condition_Gc{
					Gc: &GroupCondition{
						Logic: common.GroupLogic_And,
						Group: []*Condition{
							{
								Not: true,
								Cond: &Condition_Tc{
									Tc: &TextCondition{
										Key:  "key0",
										Term: "pattern0",
									},
								},
							},
							{
								Not: false,
								Cond: &Condition_Tc{
									Tc: &TextCondition{
										Key:  "key1",
										Term: "pattern1",
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

			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.Internal, "internal subscription storage failure"),
		},
		"empty group": {
			md: []string{
				"x-awakari-group-id", "",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
		"empty user": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-user-id in request metadata"),
		},
		"no auth info": {
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
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
				Description: k,
				Cond:        c.cond,
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
				Description: "description",
				Cond: &Condition{
					Not: false,
					Cond: &Condition_Gc{
						Gc: &GroupCondition{
							Logic: common.GroupLogic_And,
							Group: []*Condition{
								{
									Not: false,
									Cond: &Condition_Tc{
										Tc: &TextCondition{
											Key:  "key0",
											Term: "pattern0",
										},
									},
								},
								{
									Not: true,
									Cond: &Condition_Tc{
										Tc: &TextCondition{
											Key:  "key1",
											Term: "pattern1",
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
			err:  status.Error(codes.Internal, "internal subscription storage failure"),
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
				assert.Equal(t, c.sub.Description, sub.Description)
				assert.Equal(t, c.sub.Cond.Not, sub.Cond.Not)
				assert.Equal(t, c.sub.Cond.GetGc().Logic, sub.Cond.GetGc().Logic)
				assert.Equal(t, len(c.sub.Cond.GetGc().GetGroup()), len(sub.Cond.GetGc().GetGroup()))
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].Not, sub.Cond.GetGc().GetGroup()[0].Not)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetTc().Key, sub.Cond.GetGc().GetGroup()[0].GetTc().Key)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetTc().Term, sub.Cond.GetGc().GetGroup()[0].GetTc().Term)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].Not, sub.Cond.GetGc().GetGroup()[1].Not)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].GetTc().Key, sub.Cond.GetGc().GetGroup()[1].GetTc().Key)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].GetTc().Term, sub.Cond.GetGc().GetGroup()[1].GetTc().Term)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Update(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth    bool
		descr   string
		enabled bool
		err     error
	}{
		"ok1": {
			auth: true,
		},
		"ok2": {
			auth:    true,
			descr:   "new description",
			enabled: true,
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal subscription storage failure"),
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
			_, err := client.Update(ctx, &UpdateRequest{
				Id:          k,
				Description: c.descr,
				Enabled:     c.enabled,
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
			err:  status.Error(codes.Internal, "internal subscription storage failure"),
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

func TestServiceController_SearchOwn(t *testing.T) {
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
			err:  status.Error(codes.Internal, "internal subscription storage failure"),
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
		CondId: "cond0",
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
