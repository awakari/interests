package conditions_text

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestService_Create(t *testing.T) {
	//
	svc := NewService(newClientMock())
	svc = NewServiceLogging(svc, slog.Default())
	cases := map[string]struct {
		key  string
		term string
		err  error
	}{
		"ok": {
			key:  "category",
			term: "cat0",
		},
		"fail": {
			key: "fail",
			err: ErrInternal,
		},
		"conflict": {
			key: "conflict",
			err: ErrConflict,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			id, out, err := svc.Create(context.TODO(), c.key, c.term, false)
			if c.err == nil {
				assert.Equal(t, c.key, id)
				assert.Equal(t, c.term, out)
			}
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_LockCreate(t *testing.T) {
	//
	svc := NewService(newClientMock())
	svc = NewServiceLogging(svc, slog.Default())
	cases := map[string]struct {
		id  string
		err error
	}{
		"ok": {
			id: "cond0",
		},
		"missing": {
			id:  "missing",
			err: ErrNotFound,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.LockCreate(context.TODO(), c.id)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_UnlockCreate(t *testing.T) {
	//
	svc := NewService(newClientMock())
	svc = NewServiceLogging(svc, slog.Default())
	cases := map[string]struct {
		id  string
		err error
	}{
		"ok": {
			id: "cond0",
		},
		"fail": {
			id:  "fail",
			err: ErrInternal,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.UnlockCreate(context.TODO(), c.id)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_Delete(t *testing.T) {
	//
	svc := NewService(newClientMock())
	svc = NewServiceLogging(svc, slog.Default())
	cases := map[string]struct {
		id  string
		err error
	}{
		"ok": {
			id: "cond0",
		},
		"fail": {
			id:  "fail",
			err: ErrInternal,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.Delete(context.TODO(), c.id)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
