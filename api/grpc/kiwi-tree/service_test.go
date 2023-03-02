package kiwiTree

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	client := NewClientMock()
	svc := NewService(client)
	cases := []struct {
		key     string
		pattern string
		err     error
	}{
		{
			key:     "foo",
			pattern: "bar",
		},
		{
			key: "fail",
			err: ErrInternal,
		},
		{
			pattern: "locked",
			err:     ErrShouldRetry,
		},
		{
			pattern: "invalid",
			err:     ErrInvalidPatternSrc,
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s: %s", c.key, c.pattern), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Create(ctx, c.key, c.pattern)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_Delete(t *testing.T) {
	client := NewClientMock()
	svc := NewService(client)
	cases := []struct {
		key     string
		pattern string
		err     error
	}{
		{},
		{
			key: "missing",
			err: ErrNotFound,
		},
		{
			key: "fail",
			err: ErrInternal,
		},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Delete(ctx, c.key, "")
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_LockCreate(t *testing.T) {
	//
	client := NewClientMock()
	svc := NewService(client)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := svc.Create(ctx, "subject", "orders")
	require.Nil(t, err)
	err = svc.LockCreate(ctx, "subject", "orders")
	assert.Nil(t, err)
	err = svc.LockCreate(ctx, "subject", "missing")
	assert.ErrorIs(t, err, ErrNotFound)
	err = svc.LockCreate(ctx, "fail", "fail")
	assert.ErrorIs(t, err, ErrInternal)
}

func TestService_UnlockCreate(t *testing.T) {
	//
	client := NewClientMock()
	svc := NewService(client)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := svc.UnlockCreate(ctx, "key0", "orders")
	assert.Nil(t, err)
	err = svc.UnlockCreate(ctx, "key0", "fail")
	assert.ErrorIs(t, err, ErrInternal)
}
