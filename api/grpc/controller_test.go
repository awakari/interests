package grpc

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/service"
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
		excludes *MatcherGroup
		includes *MatcherGroup
		err      error
	}{
		"ok1": {},
		"ok2": {
			excludes: &MatcherGroup{
				Matchers: []*Matcher{
					{
						Partial: true,
						Key:     "key0",
						Pattern: &Pattern{
							Code: []byte("pattern0"),
							Src:  "pattern0",
						},
					},
				},
			},
			includes: &MatcherGroup{
				All: true,
				Matchers: []*Matcher{
					{
						Key: "key0",
						Pattern: &Pattern{
							Code: []byte("pattern0"),
							Src:  "pattern0",
						},
					},
				},
			},
		},
		"fail": {
			err: status.Error(codes.Internal, "internal failure"),
		},
		"invalid": {
			err: status.Error(codes.InvalidArgument, "invalid subscription"),
		},
		"conflict": {
			err: status.Error(codes.AlreadyExists, "subscription already exists"),
		},
		"busy": {
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
				Excludes: c.excludes,
				Includes: c.includes,
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
				Includes: &MatcherGroup{
					All: true,
					Matchers: []*Matcher{
						{
							Partial: true,
							Key:     "key0",
							Pattern: &Pattern{
								Code: []byte("pattern0"),
								Src:  "pattern0",
							},
						},
					},
				},
				Excludes: &MatcherGroup{
					Matchers: []*Matcher{
						{
							Key: "key1",
							Pattern: &Pattern{
								Code: []byte("pattern1"),
								Src:  "pattern1",
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
				assert.Equal(t, c.sub.Excludes.All, sub.Excludes.All)
				assert.Equal(t, c.sub.Excludes.Matchers[0].Key, sub.Excludes.Matchers[0].Key)
				assert.Equal(t, c.sub.Excludes.Matchers[0].Partial, sub.Excludes.Matchers[0].Partial)
				assert.Equal(t, c.sub.Excludes.Matchers[0].Pattern.Code, sub.Excludes.Matchers[0].Pattern.Code)
				assert.Equal(t, c.sub.Excludes.Matchers[0].Pattern.Src, sub.Excludes.Matchers[0].Pattern.Src)
				assert.Equal(t, c.sub.Includes.All, sub.Includes.All)
				assert.Equal(t, c.sub.Includes.Matchers[0].Key, sub.Includes.Matchers[0].Key)
				assert.Equal(t, c.sub.Includes.Matchers[0].Partial, sub.Includes.Matchers[0].Partial)
				assert.Equal(t, c.sub.Includes.Matchers[0].Pattern.Code, sub.Includes.Matchers[0].Pattern.Code)
				assert.Equal(t, c.sub.Includes.Matchers[0].Pattern.Src, sub.Includes.Matchers[0].Pattern.Src)
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
			resp, err := client.Search(ctx, &SearchRequest{Cursor: k})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.count, len(resp.Page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
