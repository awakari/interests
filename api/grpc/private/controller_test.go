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
	"io"
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
