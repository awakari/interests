package service

import (
	"context"
	"fmt"
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
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
	)
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 4",
			CreateRequest{
				Description: "pre existing",
				Routes: []string{
					"route 4",
				},
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
				Routes: []string{
					"route",
				},
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
				Routes: []string{
					"route 1",
				},
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
				Routes: []string{
					"route 2",
				},
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
				Routes: []string{
					"route 3",
				},
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
				Routes: []string{
					"route 4",
				},
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
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
	)
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 1",
			CreateRequest{
				Description: "pre existing",
				Routes: []string{
					"route 1",
				},
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
				Routes: []string{
					"route 1",
				},
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
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
	)
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 1",
			CreateRequest{
				Description: "pre existing",
				Routes: []string{
					"route 1",
				},
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
	require.Nil(
		t, svc.Create(
			nil,
			"subscription 2",
			CreateRequest{
				Description: "fails to clean up matchers",
				Routes: []string{
					"route 2",
				},
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
									Src: "fail",
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
		err    error
		errMsg string
	}{
		"subscription 0": {
			err:    ErrNotFound,
			errMsg: "subscription was not found: subscription was not found by name: subscription 0",
		},
		"subscription 1": {},
		"subscription 2": {
			err:    ErrCleanMatcher,
			errMsg: "matchers cleanup failure, may cause matchers garbage: internal failure, subscription: {subscription 2 fails to clean up matchers [route 2] {false [{{key0 pattern0} false}]} {false [{{key1 fail} true}]}}",
		},
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
				assert.Equal(t, c.errMsg, err.Error())
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
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
	)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	//
	for i := 0; i < 5; i++ {
		req := CreateRequest{
			Description: "my subscription",
			Routes: []string{
				"route",
			},
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

func TestService_Search(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	excCompleteMatchersSvc := matchers.NewServiceMock()
	excPartialMatchersSvc := matchers.NewServiceMock()
	incCompleteMatchersSvc := matchers.NewServiceMock()
	incPartialMatchersSvc := matchers.NewServiceMock()
	svc := NewService(
		stor,
		excCompleteMatchersSvc,
		excPartialMatchersSvc,
		incCompleteMatchersSvc,
		incPartialMatchersSvc,
	)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		inExcludes := i%2 == 1 // every 2nd: 1, 3, 5, 7, 9, ...
		matchers := []model.Matcher{
			{
				Partial: i%3 == 2, // every 3rd: 2, 5, 8, ...
				MatcherData: model.MatcherData{
					Key: fmt.Sprintf("key%d", i%4),
					Pattern: model.Pattern{
						Src: fmt.Sprintf("pattern%d", i%5),
					},
				},
			},
		}
		req := CreateRequest{
			Routes: []string{
				"route 4",
			},
		}
		if inExcludes {
			req.Excludes = model.MatcherGroup{
				Matchers: matchers,
			}
		} else {
			req.Includes = model.MatcherGroup{
				Matchers: matchers,
			}
		}
		require.Nil(t, svc.Create(ctx, fmt.Sprintf("sub%d", i), req))
	}
	//
	cases := map[string]struct {
		query  Query
		cursor string
		page   []model.Subscription
		err    error
	}{
		"key0/pattern0 -> 3 subs": {
			query: Query{
				Limit: 10,
				Matcher: model.Matcher{
					MatcherData: model.MatcherData{
						Key: "key0",
						Pattern: model.Pattern{
							Code: []byte("pattern0"),
						},
					},
				},
			},
			page: []model.Subscription{
				{
					Name: "sub0",
					Routes: []string{
						"route 4",
					},
					Includes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key0",
									Pattern: model.Pattern{
										Code: []byte("pattern0"),
										Src:  "pattern0",
									},
								},
							},
						},
					},
				},
				{
					Name: "sub40",
					Routes: []string{
						"route 4",
					},
					Includes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key0",
									Pattern: model.Pattern{
										Code: []byte("pattern0"),
										Src:  "pattern0",
									},
								},
							},
						},
					},
				},
				{
					Name: "sub60",
					Routes: []string{
						"route 4",
					},
					Includes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key0",
									Pattern: model.Pattern{
										Code: []byte("pattern0"),
										Src:  "pattern0",
									},
								},
							},
						},
					},
				},
			},
		},
		"key1/pattern1, limit=2": {
			query: Query{
				Limit:      2,
				InExcludes: true,
				Matcher: model.Matcher{
					MatcherData: model.MatcherData{
						Key: "key1",
						Pattern: model.Pattern{
							Code: []byte("pattern1"),
						},
					},
				},
			},
			page: []model.Subscription{
				{
					Name: "sub1",
					Routes: []string{
						"route 4",
					},
					Excludes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key1",
									Pattern: model.Pattern{
										Code: []byte("pattern1"),
										Src:  "pattern1",
									},
								},
							},
						},
					},
				},
				{
					Name: "sub21",
					Routes: []string{
						"route 4",
					},
					Excludes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key1",
									Pattern: model.Pattern{
										Code: []byte("pattern1"),
										Src:  "pattern1",
									},
								},
							},
						},
					},
				},
			},
		},
		"key1/pattern1, cursor=sub21": {
			query: Query{
				Limit:      3,
				InExcludes: true,
				Matcher: model.Matcher{
					MatcherData: model.MatcherData{
						Key: "key1",
						Pattern: model.Pattern{
							Code: []byte("pattern1"),
						},
					},
				},
			},
			cursor: "sub21",
			page: []model.Subscription{
				{
					Name: "sub61",
					Routes: []string{
						"route 4",
					},
					Excludes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key1",
									Pattern: model.Pattern{
										Code: []byte("pattern1"),
										Src:  "pattern1",
									},
								},
							},
						},
					},
				},
				{
					Name: "sub81",
					Routes: []string{
						"route 4",
					},
					Excludes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								MatcherData: model.MatcherData{
									Key: "key1",
									Pattern: model.Pattern{
										Code: []byte("pattern1"),
										Src:  "pattern1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			page, err := svc.Search(ctx, c.query, c.cursor)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, len(c.page), len(page))
				for i, sub := range page {
					assert.Equal(t, c.page[i], sub)
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
