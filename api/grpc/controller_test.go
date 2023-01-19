package grpc

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"os"
	"testing"
	"time"
)

const (
	port = 8080
)

var (
	log = slog.Default()
)

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
		cond *InputCondition
		err  error
	}{
		"ok1": {
			cond: &InputCondition{
				Condition: &InputCondition_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeCondition{
						Base: &KiwiCondition{
							Base: &KeyCondition{
								Base: &ConditionBase{},
							},
						},
					},
				},
			},
		},
		"ok2": {
			cond: &InputCondition{
				Condition: &InputCondition_GroupCondition{
					GroupCondition: &GroupInputCondition{
						Base: &GroupConditionBase{
							Base: &ConditionBase{
								Not: false,
							},
							Logic: GroupLogic_And,
						},
						Group: []*InputCondition{
							{
								Condition: &InputCondition_KiwiTreeCondition{
									KiwiTreeCondition: &KiwiTreeCondition{
										Base: &KiwiCondition{
											Base: &KeyCondition{
												Base: &ConditionBase{
													Not: true,
												},
												Key: "key0",
											},
											Pattern: "pattern0",
											Partial: true,
										},
									},
								},
							},
							{
								Condition: &InputCondition_KiwiTreeCondition{
									KiwiTreeCondition: &KiwiTreeCondition{
										Base: &KiwiCondition{
											Base: &KeyCondition{
												Base: &ConditionBase{
													Not: false,
												},
												Key: "key1",
											},
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
			cond: &InputCondition{
				Condition: &InputCondition_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeCondition{
						Base: &KiwiCondition{
							Base: &KeyCondition{
								Base: &ConditionBase{},
							},
						},
					},
				},
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"invalid": {
			cond: &InputCondition{
				Condition: &InputCondition_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeCondition{
						Base: &KiwiCondition{
							Base: &KeyCondition{
								Base: &ConditionBase{},
							},
						},
					},
				},
			},
			err: status.Error(codes.InvalidArgument, "invalid subscription"),
		},
		"conflict": {
			cond: &InputCondition{
				Condition: &InputCondition_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeCondition{
						Base: &KiwiCondition{
							Base: &KeyCondition{
								Base: &ConditionBase{},
							},
						},
					},
				},
			},
			err: status.Error(codes.AlreadyExists, "subscription already exists"),
		},
		"busy": {
			cond: &InputCondition{
				Condition: &InputCondition_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeCondition{
						Base: &KiwiCondition{
							Base: &KeyCondition{
								Base: &ConditionBase{},
							},
						},
					},
				},
			},
			err: status.Error(codes.Unavailable, "retry the operation"),
		},
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err = client.Create(ctx, &CreateRequest{
				Name:        k,
				Description: k,
				Routes: []string{
					"destination",
				},
				Condition: c.cond,
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
		sub *Subscription
		err error
	}{
		"ok": {
			sub: &Subscription{
				Name:        "ok",
				Description: "description",
				Routes: []string{
					"destination",
				},
				Condition: &OutputCondition{
					Condition: &OutputCondition_GroupCondition{
						GroupCondition: &GroupOutputCondition{
							Base: &GroupConditionBase{
								Base: &ConditionBase{
									Not: false,
								},
								Logic: GroupLogic_And,
							},
							Group: []*OutputCondition{
								{
									Condition: &OutputCondition_KiwiCondition{
										KiwiCondition: &KiwiCondition{
											Base: &KeyCondition{
												Base: &ConditionBase{
													Not: false,
												},
												Key: "key0",
											},
											Pattern: "pattern0",
											Partial: true,
										},
									},
								},
								{
									Condition: &OutputCondition_KiwiCondition{
										KiwiCondition: &KiwiCondition{
											Base: &KeyCondition{
												Base: &ConditionBase{
													Not: true,
												},
												Key: "key1",
											},
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
			err: status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			err: status.Error(codes.NotFound, "subscription was not found"),
		},
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			sub, err := client.Read(ctx, &ReadRequest{Name: k})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sub.Name, sub.Name)
				assert.Equal(t, c.sub.Description, sub.Description)
				assert.Equal(t, c.sub.Routes, sub.Routes)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().Base.Base.Not, sub.Condition.GetGroupCondition().Base.Base.Not)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().Base.Logic, sub.Condition.GetGroupCondition().Base.Logic)
				assert.Equal(t, len(c.sub.Condition.GetGroupCondition().GetGroup()), len(sub.Condition.GetGroupCondition().GetGroup()))
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Base.Base.Not, sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Base.Base.Not)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Base.Key, sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Base.Key)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Pattern, sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Pattern)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Partial, sub.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Partial)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Base.Base.Not, sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Base.Base.Not)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Base.Key, sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Base.Key)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Pattern, sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Pattern)
				assert.Equal(t, c.sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Partial, sub.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Partial)
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
		sub *Subscription
		err error
	}{
		"ok": {},
		"fail": {
			err: status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			err: status.Error(codes.NotFound, "subscription was not found"),
		},
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := client.Delete(ctx, &DeleteRequest{Name: k})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_ListNames(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		err   error
		names []string
	}{
		"": {
			names: []string{
				"sub0",
				"sub1",
			},
		},
		"fail": {
			err: status.Error(codes.Internal, "internal failure"),
		},
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.ListNames(ctx, &ListNamesRequest{Cursor: k})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, len(c.names), len(resp.Names))
				assert.ElementsMatch(t, c.names, resp.Names)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Search(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		err   error
		count int
	}{
		"": {
			count: 2,
		},
		"fail": {
			err: status.Error(codes.Internal, "internal failure"),
		},
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.SearchByKiwi(ctx, &SearchByKiwiRequest{Cursor: k})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.count, len(resp.Page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
