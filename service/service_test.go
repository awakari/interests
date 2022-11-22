package service

import (
	"context"
	"github.com/meandros-messaging/subscriptions/model"
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
