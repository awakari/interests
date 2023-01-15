package kiwi

import (
	"context"
	"fmt"
	grpcApi "github.com/awakari/subscriptions/api/grpc/kiwi-tree"
	"github.com/awakari/subscriptions/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	client := grpcApi.NewClientMock()
	svc := NewService(client)
	cases := []struct {
		key        string
		patternSrc string
		m          model.MatcherData
		err        error
	}{
		{
			key:        "foo",
			patternSrc: "bar",
			m: model.MatcherData{
				Key: "foo",
				Pattern: model.Pattern{
					Code: []byte("bar"),
					Src:  "bar",
				},
			},
		},
		{
			key: "fail",
			err: ErrInternal,
		},
		{
			patternSrc: "locked",
			err:        ErrShouldRetry,
		},
		{
			patternSrc: "invalid",
			err:        ErrInvalidPatternSrc,
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s: %s", c.key, c.patternSrc), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			m, err := svc.Create(ctx, c.key, c.patternSrc)
			if c.err != nil {
				assert.ErrorIs(t, err, c.err)
			} else {
				assert.Equal(t, c.m, m)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	client := grpcApi.NewClientMock()
	svc := NewService(client)
	cases := []struct {
		key        string
		patternSrc string
		m          model.MatcherData
		err        error
	}{
		{
			m: model.MatcherData{
				Key: "foo",
				Pattern: model.Pattern{
					Code: []byte("bar"),
					Src:  "bar",
				},
			},
		},
		{
			m: model.MatcherData{
				Key: "missing",
			},
			err: ErrNotFound,
		},
		{
			m: model.MatcherData{
				Key: "fail",
			},
			err: ErrInternal,
		},
	}
	for _, c := range cases {
		t.Run(c.m.Key+": "+c.m.Pattern.String(), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Delete(ctx, c.m)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_LockCreate(t *testing.T) {
	//
	client := grpcApi.NewClientMock()
	svc := NewService(client)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := svc.Create(ctx, "subject", "orders")
	require.Nil(t, err)
	err = svc.LockCreate(ctx, []byte("orders"))
	assert.Nil(t, err)
	err = svc.LockCreate(ctx, []byte("missing"))
	assert.ErrorIs(t, err, ErrNotFound)
	err = svc.LockCreate(ctx, []byte("fail"))
	assert.ErrorIs(t, err, ErrInternal)
}

func TestService_UnlockCreate(t *testing.T) {
	//
	client := grpcApi.NewClientMock()
	svc := NewService(client)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := svc.UnlockCreate(ctx, []byte("orders"))
	assert.Nil(t, err)
	err = svc.UnlockCreate(ctx, []byte("fail"))
	assert.ErrorIs(t, err, ErrInternal)
}
