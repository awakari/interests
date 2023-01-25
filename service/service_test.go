package service

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model"
	kiwiTree "github.com/awakari/subscriptions/service/kiwi-tree"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	//
	storMem := make(map[string]model.Subscription)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	require.Nil(
		t, svc.Create(
			nil,
			model.Subscription{
				Name:        "subscription 4",
				Description: "pre existing",
				Routes: []string{
					"route 4",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"key0",
						),
						false,
						"pattern0",
					),
				),
			},
		),
	)
	//
	cases := map[string]struct {
		req model.Subscription
		err error
	}{
		"empty": {
			req: model.Subscription{
				Name:        "subscription 0",
				Description: "my subscription",
			},
			err: model.ErrInvalidSubscription,
		},
		"empty name": {
			req: model.Subscription{
				Description: "my subscription",
				Routes: []string{
					"route",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"",
						),
						false,
						"ok",
					),
				),
			},
			err: model.ErrInvalidSubscription,
		},
		"locked": {
			req: model.Subscription{
				Name:        "subscription 1",
				Description: "my subscription",
				Routes: []string{
					"route 1",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"",
						),
						false,
						"locked",
					),
				),
			},
			err: ErrShouldRetry,
		},
		"fail": {
			req: model.Subscription{
				Name:        "subscription 2",
				Description: "my subscription",
				Routes: []string{
					"route 2",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"fail",
						),
						false,
						"fail",
					),
				),
			},
			err: ErrInternal,
		},
		"ok": {
			req: model.Subscription{
				Name:        "subscription 3",
				Description: "my subscription",
				Routes: []string{
					"route 3",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"key0",
						),
						false,
						"ok",
					),
				),
			},
		},
		"conflict": {
			req: model.Subscription{
				Name: "subscription 4",
				Routes: []string{
					"route 4",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"key0",
						),
						false,
						"pattern0",
					),
				),
			},
			err: ErrConflict,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Create(ctx, c.req)
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
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	require.Nil(
		t, svc.Create(
			nil,
			model.Subscription{
				Name:        "subscription 1",
				Description: "pre existing",
				Routes: []string{
					"route 1",
				},
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"key0",
						),
						false,
						"pattern0",
					),
				),
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
				Condition: model.NewKiwiTreeCondition(
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"key0",
						),
						false,
						"pattern0",
					),
				),
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
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	require.Nil(
		t, svc.Create(
			nil,
			model.Subscription{
				Name:        "subscription 1",
				Description: "pre existing",
				Routes: []string{
					"route 1",
				},
				Condition: model.NewGroupCondition(
					model.NewCondition(false),
					model.GroupLogicAnd,
					[]model.Condition{
						model.NewKiwiTreeCondition(
							model.NewKiwiCondition(
								model.NewKeyCondition(
									model.NewCondition(false),
									"key0",
								),
								false,
								"pattern0",
							),
						),
						model.NewKiwiTreeCondition(
							model.NewKiwiCondition(
								model.NewKeyCondition(
									model.NewCondition(true),
									"key1",
								),
								true,
								"pattern1",
							),
						),
					},
				),
			},
		),
	)
	require.Nil(
		t, svc.Create(
			nil,
			model.Subscription{
				Name:        "subscription 2",
				Description: "fails to clean up kiwis",
				Routes: []string{
					"route 2",
				},
				Condition: model.NewGroupCondition(
					model.NewCondition(false),
					model.GroupLogicAnd,
					[]model.Condition{
						model.NewKiwiTreeCondition(
							model.NewKiwiCondition(
								model.NewKeyCondition(
									model.NewCondition(false),
									"key0",
								),
								false,
								"pattern0",
							),
						),
						model.NewKiwiTreeCondition(
							model.NewKiwiCondition(
								model.NewKeyCondition(
									model.NewCondition(true),
									"key1",
								),
								true,
								"fail",
							),
						),
					},
				),
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
			err:    ErrCleanKiwis,
			errMsg: "kiwis cleanup failure, may cause kiwis garbage: internal failure, subscription: {subscription 2 fails to clean up kiwis [route 2] {{false} And [{{{{false} key0} false pattern0}} {{{{true} key1} true fail}}]}}",
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
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	//
	for i := 0; i < 5; i++ {
		req := model.Subscription{
			Name:        fmt.Sprintf("sub%d", i),
			Description: "my subscription",
			Routes: []string{
				"route",
			},
			Condition: model.NewKiwiTreeCondition(
				model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						"key0",
					),
					false,
					"pattern0",
				),
			),
		}

		require.Nil(t, svc.Create(ctx, req))
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
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := model.Subscription{
			Name: fmt.Sprintf("sub%d", i),
			Routes: []string{
				"route 4",
			},
			Condition: model.NewKiwiTreeCondition(
				model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						fmt.Sprintf("key%d", i%4),
					),
					i%3 == 2,
					fmt.Sprintf("pattern%d", i%5),
				),
			),
		}
		require.Nil(t, svc.Create(ctx, req))
	}
	//
	cases := map[string]struct {
		query  model.ConditionQuery
		cursor string
		page   []model.Subscription
		err    error
	}{
		"key0/pattern0 -> 3 subs": {
			query: model.ConditionQuery{
				Limit: 10,
				Condition: model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						"key0",
					),
					false,
					"pattern0",
				),
			},
			page: []model.Subscription{
				{
					Name: "sub0",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key0"),
							),
							false,
							"pattern0",
						),
					),
				},
				{
					Name: "sub20",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key0"),
							),
							true,
							"pattern0",
						),
					),
				},
				{
					Name: "sub40",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key0"),
							),
							false,
							"pattern0",
						),
					),
				},
				{
					Name: "sub60",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key0"),
							),
							false,
							"pattern0",
						),
					),
				},
				{
					Name: "sub80",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key0"),
							),
							true,
							"pattern0",
						),
					),
				},
			},
		},
		"key1/pattern1, limit=2": {
			query: model.ConditionQuery{
				Limit: 2,
				Condition: model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						"key1",
					),
					false,
					"pattern1",
				),
			},
			page: []model.Subscription{
				{
					Name: "sub1",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key1"),
							),
							false,
							"pattern1",
						),
					),
				},
				{
					Name: "sub21",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key1"),
							),
							false,
							"pattern1",
						),
					),
				},
			},
		},
		"key1/pattern1, cursor=sub21": {
			query: model.ConditionQuery{
				Limit: 3,
				Condition: model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						"key1",
					),
					false,
					"pattern1",
				),
			},
			cursor: "sub21",
			page: []model.Subscription{
				{
					Name: "sub41",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key1"),
							),
							true,
							"pattern1",
						),
					),
				},
				{
					Name: "sub61",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key1"),
							),
							false,
							"pattern1",
						),
					),
				},
				{
					Name: "sub81",
					Routes: []string{
						"route 4",
					},
					Condition: model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								fmt.Sprintf("key1"),
							),
							false,
							"pattern1",
						),
					),
				},
			},
		},
		"fail on group condition query": {
			query: model.ConditionQuery{
				Limit: 3,
				Condition: model.NewGroupCondition(
					model.NewCondition(false),
					model.GroupLogicAnd,
					[]model.Condition{},
				),
			},
			err: ErrInvalidQuery,
		},
		"fail on base condition query": {
			query: model.ConditionQuery{
				Limit:     3,
				Condition: model.NewCondition(false),
			},
			err: ErrInvalidQuery,
		},
		"fail on key condition query": {
			query: model.ConditionQuery{
				Limit: 3,
				Condition: model.NewKeyCondition(
					model.NewCondition(false),
					"key0",
				),
			},
			err: ErrInvalidQuery,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			page, err := svc.SearchByCondition(ctx, c.query, c.cursor)
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
