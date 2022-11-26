package service

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/aggregator"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	svc := NewService(
		stor,
		10,
		nil,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
		0,
		nil,
	)
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 4",
			CreateRequest{
				Description: "pre existing",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Key: "key0",
								Pattern: model.Pattern{
									Src: "pattern0",
								},
							},
						},
					},
				},
			},
		),
	)
	//
	cases := map[string]struct {
		name string
		req  CreateRequest
		err  error
	}{
		"empty": {
			name: "subscription 0",
			req: CreateRequest{
				Description: "my subscription",
			},
			err: model.ErrInvalidSubscription,
		},
		"empty name": {
			name: "",
			req: CreateRequest{
				Description: "my subscription",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Pattern: model.Pattern{
									Src: "ok",
								},
							},
						},
					},
				},
			},
			err: model.ErrInvalidSubscription,
		},
		"locked": {
			name: "subscription 1",
			req: CreateRequest{
				Description: "my subscription",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Pattern: model.Pattern{
									Src: "locked",
								},
							},
						},
					},
				},
			},
			err: ErrShouldRetry,
		},
		"fail": {
			name: "subscription 2",
			req: CreateRequest{
				Description: "my subscription",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Key: "fail",
							},
						},
					},
				},
			},
			err: ErrInternal,
		},
		"ok": {
			name: "subscription 3",
			req: CreateRequest{
				Description: "my subscription",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							Partial: true,
							MatcherData: model.MatcherData{
								Key: "ok",
							},
						},
					},
				},
			},
		},
		"conflict": {
			name: "subscription 4",
			req: CreateRequest{
				Excludes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							Partial: true,
							MatcherData: model.MatcherData{
								Key: "key0",
								Pattern: model.Pattern{
									Src: "pattern0",
								},
							},
						},
					},
				},
			},
			err: ErrConflict,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Create(ctx, c.name, c.req)
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_Read(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	svc := NewService(
		stor,
		10,
		nil,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
		0,
		nil,
	)
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 1",
			CreateRequest{
				Description: "pre existing",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Key: "key0",
								Pattern: model.Pattern{
									Src: "pattern0",
								},
							},
						},
					},
				},
			},
		),
	)
	//
	cases := map[string]struct {
		sub model.Subscription
		err error
	}{
		"subscription 0": {
			err: ErrNotFound,
		},
		"subscription 1": {
			sub: model.Subscription{
				Name:        "subscription 1",
				Description: "pre existing",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Key: "key0",
								Pattern: model.Pattern{
									Src:  "pattern0",
									Code: model.PatternCode{0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x30},
								},
							},
						},
					},
				},
				Excludes: model.MatcherGroup{},
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			sub, err := svc.Read(ctx, name)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sub, sub)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	svc := NewService(
		stor,
		10,
		nil,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
		0,
		nil,
	)
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 1",
			CreateRequest{
				Description: "pre existing",
				Includes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							MatcherData: model.MatcherData{
								Key: "key0",
								Pattern: model.Pattern{
									Src: "pattern0",
								},
							},
						},
					},
				},
				Excludes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							Partial: true,
							MatcherData: model.MatcherData{
								Key: "key1",
								Pattern: model.Pattern{
									Src: "pattern1",
								},
							},
						},
					},
				},
			},
		),
	)
	//
	cases := map[string]struct {
		err error
	}{
		"subscription 0": {
			ErrNotFound,
		},
		"subscription 1": {},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Delete(ctx, name)
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_ListNames(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	svc := NewService(
		stor,
		10,
		nil,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
		0,
		nil,
	)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	//
	for i := 0; i < 5; i++ {
		req := CreateRequest{
			Description: "my subscription",
			Includes: model.MatcherGroup{
				Matchers: []model.Matcher{
					{
						MatcherData: model.MatcherData{
							Key: "ok",
						},
					},
				},
			},
		}

		require.Nil(t, svc.Create(ctx, fmt.Sprintf("sub%d", i), req))
	}
	//
	cases := map[string]struct {
		err    error
		result []string
	}{
		"": {
			result: []string{
				"sub0",
				"sub1",
				"sub2",
				"sub3",
				"sub4",
			},
		},
		"fail": {
			err: ErrInternal,
		},
	}
	//
	for cursor, c := range cases {
		t.Run(cursor, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			p, err := svc.ListNames(ctx, 0, cursor)
			if c.err != nil {
				assert.ErrorIs(t, err, c.err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, len(c.result), len(p))
				for i := 0; i < len(p); i++ {
					assert.Equal(t, c.result[i], p[i])
				}
			}
		})
	}
}

func TestService_resolveSubscriptions(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	aggregatorSink := make(chan aggregator.Match, 10)
	aggregatorSvc := aggregator.NewServiceMock(aggregatorSink)
	svc := NewService(
		stor,
		10,
		nil,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
		10,
		aggregatorSvc,
	)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	//
	for i := 0; i < 10; i++ {
		req := CreateRequest{
			Description: fmt.Sprintf("my subscription #%d", i),
			Includes: model.MatcherGroup{
				Matchers: []model.Matcher{
					{
						MatcherData: model.MatcherData{
							Key: fmt.Sprintf("key%d", i),
						},
					},
				},
			},
		}
		require.Nil(t, svc.Create(ctx, fmt.Sprintf("sub%d", i), req))
	}
	//
	err := svc.(service).resolveSubscriptions(
		ctx,
		31415926,
		model.Matcher{
			MatcherData: model.MatcherData{
				Key: "key1",
			},
		},
		false,
	)
	assert.Nil(t, err)
	select {
	case m := <-aggregatorSink:
		assert.Equal(t, aggregator.Match{
			MessageId:        31415926,
			SubscriptionName: "sub1",
			Includes: aggregator.MatchGroup{
				MatcherCount: 1,
			},
		}, m)
	default:
		assert.Fail(t, "no match received")
	}
}
