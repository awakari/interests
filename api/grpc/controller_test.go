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
		cond *ConditionInput
		err  error
	}{
		"ok1": {
			cond: &ConditionInput{
				Condition: &ConditionInput_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeConditionInput{},
				},
			},
		},
		"ok2": {
			cond: &ConditionInput{
				Not: false,
				Condition: &ConditionInput_GroupCondition{
					GroupCondition: &GroupConditionInput{
						Logic: GroupLogic_And,
						Group: []*ConditionInput{
							{
								Not: true,
								Condition: &ConditionInput_KiwiTreeCondition{
									KiwiTreeCondition: &KiwiTreeConditionInput{
										Key:     "key0",
										Pattern: "pattern0",
										Partial: true,
									},
								},
							},
							{
								Not: false,
								Condition: &ConditionInput_KiwiTreeCondition{
									KiwiTreeCondition: &KiwiTreeConditionInput{
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
			cond: &ConditionInput{
				Condition: &ConditionInput_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"invalid": {
			cond: &ConditionInput{
				Condition: &ConditionInput_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.InvalidArgument, "invalid subscription route"),
		},
		"conflict": {
			cond: &ConditionInput{
				Condition: &ConditionInput_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeConditionInput{},
				},
			},
			err: status.Error(codes.AlreadyExists, "subscription already exists"),
		},
		"busy": {
			cond: &ConditionInput{
				Condition: &ConditionInput_KiwiTreeCondition{
					KiwiTreeCondition: &KiwiTreeConditionInput{},
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
			_, err = client.Create(ctx, &SubscriptionDataInput{
				Metadata: map[string]string{
					"description": k,
				},
				Route: &RouteInput{
					Destinations: []string{
						"destination",
					},
					Condition: c.cond,
				},
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
		sub *SubscriptionDataOutput
		err error
	}{
		"ok": {
			sub: &SubscriptionDataOutput{
				Metadata: map[string]string{
					"description": "description",
				},
				Route: &RouteOutput{
					Destinations: []string{
						"destination",
					},
					Condition: &ConditionOutput{
						Not: false,
						Condition: &ConditionOutput_GroupCondition{
							GroupCondition: &GroupConditionOutput{
								Logic: GroupLogic_And,
								Group: []*ConditionOutput{
									{
										Not: false,
										Condition: &ConditionOutput_KiwiCondition{
											KiwiCondition: &KiwiConditionOutput{
												Key:     "key0",
												Pattern: "pattern0",
												Partial: true,
											},
										},
									},
									{
										Not: true,
										Condition: &ConditionOutput_KiwiCondition{
											KiwiCondition: &KiwiConditionOutput{
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
			sub, err := client.Read(ctx, &ReadRequest{Id: k})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sub.Metadata, sub.Metadata)
				assert.Equal(t, c.sub.Route.Destinations, sub.Route.Destinations)
				assert.Equal(t, c.sub.Route.Condition.Not, sub.Route.Condition.Not)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().Logic, sub.Route.Condition.GetGroupCondition().Logic)
				assert.Equal(t, len(c.sub.Route.Condition.GetGroupCondition().GetGroup()), len(sub.Route.Condition.GetGroupCondition().GetGroup()))
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[0].Not, sub.Route.Condition.GetGroupCondition().GetGroup()[0].Not)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Key, sub.Route.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Key)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Pattern, sub.Route.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Pattern)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Partial, sub.Route.Condition.GetGroupCondition().GetGroup()[0].GetKiwiCondition().Partial)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[1].Not, sub.Route.Condition.GetGroupCondition().GetGroup()[1].Not)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Key, sub.Route.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Key)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Pattern, sub.Route.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Pattern)
				assert.Equal(t, c.sub.Route.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Partial, sub.Route.Condition.GetGroupCondition().GetGroup()[1].GetKiwiCondition().Partial)
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
			_, err := client.Delete(ctx, &DeleteRequest{Id: k})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_SearchByCondition(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		condition isSearchByConditionRequest_Condition
		err       error
		count     int
	}{
		"": {
			condition: &SearchByConditionRequest_KiwiConditionQuery{
				KiwiConditionQuery: &KiwiConditionQuery{},
			},
			count: 2,
		},
		"fail": {
			err: status.Error(codes.InvalidArgument, "unsupported condition type"),
		},
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.SearchByCondition(
				ctx,
				&SearchByConditionRequest{
					Condition: c.condition,
					Cursor:    k,
				},
			)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.count, len(resp.Page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_SearchByMetadata(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		metadata map[string]string
		err      error
		count    int
	}{
		"": {
			metadata: map[string]string{},
			count:    2,
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
			resp, err := client.SearchByMetadata(
				ctx,
				&SearchByMetadataRequest{
					Metadata: c.metadata,
					Cursor:   k,
				},
			)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.count, len(resp.Page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
