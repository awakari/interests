package private

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

const port = 8081

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

func TestServiceController_SearchByCondition(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		condition isSearchByConditionRequest_Cond
		err       error
		count     int
	}{
		"": {
			condition: &SearchByConditionRequest_Kcq{
				Kcq: &KiwiConditionQuery{},
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
					Cond: c.condition,
					Cursor: &ConditionMatchKey{
						SubId: k,
					},
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
