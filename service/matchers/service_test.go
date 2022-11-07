package matchers

import (
	"context"
	"fmt"
	grpcApi "github.com/meandros-messaging/subscriptions/api/grpc/matchers"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/stretchr/testify/assert"
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

func TestService_Search(t *testing.T) {
	client := grpcApi.NewClientMock()
	svc := NewService(client)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	page, err := svc.Search(ctx, "foo", "bar", 123, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(page))
	assert.Equal(t, model.PatternCode("abc"), page[0])
	assert.Equal(t, model.PatternCode("def"), page[1])
}

func TestService_Search_Fail(t *testing.T) {
	client := grpcApi.NewClientMock()
	svc := NewService(client)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	page, err := svc.Search(ctx, "fail", "bar", 123, nil)
	assert.ErrorIs(t, err, ErrInternal)
	assert.Empty(t, page)
}
