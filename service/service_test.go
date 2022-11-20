package service

import (
	"context"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/aggregator"
	"github.com/meandros-messaging/subscriptions/service/lexemes"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	//
	storMem := make(map[model.SubscriptionKey]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	lexemesSvc := lexemes.NewServiceMock()
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	aggregatorExcStor := make(map[model.MessageId]map[model.SubscriptionKey]aggregator.MatchInGroupMock)
	aggregatorIncStor := make(map[model.MessageId]map[model.SubscriptionKey]aggregator.MatchInGroupMock)
	aggregatorSvc := aggregator.NewServiceMock(aggregatorIncStor, aggregatorExcStor)
	svc := NewService(
		stor,
		10,
		lexemesSvc,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
		0,
		aggregatorSvc,
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
