package service

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	kiwiTree "github.com/awakari/subscriptions/service/kiwi-tree"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	_, err := svc.Create(
		nil,
		subscription.Data{
			Metadata: map[string]string{
				"description": "pre-existing",
			},
			Route: subscription.Route{
				Destinations: []string{
					"route 4",
				},
				Condition: condition.NewKiwiTreeCondition(
					condition.NewKiwiCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
						false,
						"pattern0",
					),
				),
			},
		},
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		req subscription.Data
		err error
	}{
		"empty": {
			err: subscription.ErrInvalidSubscriptionRoute,
		},
		"locked": {
			req: subscription.Data{
				Metadata: map[string]string{
					"description": "my subscription",
				},
				Route: subscription.Route{
					Destinations: []string{
						"route 1",
					},
					Condition: condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "", ""),
							false,
							"locked",
						),
					),
				},
			},
			err: ErrShouldRetry,
		},
		"fail": {
			req: subscription.Data{
				Metadata: map[string]string{
					"description": "my subscription",
				},
				Route: subscription.Route{
					Destinations: []string{
						"route 2",
					},
					Condition: condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(
								condition.NewCondition(false),
								"", "fail",
							),
							false,
							"fail",
						),
					),
				},
			},
			err: ErrInternal,
		},
		"ok": {
			req: subscription.Data{
				Metadata: map[string]string{
					"description": "my subscription",
				},
				Route: subscription.Route{
					Destinations: []string{
						"route 3",
					},
					Condition: condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
							false,
							"ok",
						),
					),
				},
			},
		},
		"conflict": {
			req: subscription.Data{
				Metadata: map[string]string{
					"description": "conflict",
				},
				Route: subscription.Route{
					Destinations: []string{
						"route 4",
					},
					Condition: condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(
								condition.NewCondition(false),
								"", "key0",
							),
							false,
							"pattern0",
						),
					),
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
			id, err := svc.Create(ctx, c.req)
			if c.err == nil {
				assert.Nil(t, err)
				assert.NotEmpty(t, id)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_Read(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	id1, err := svc.Create(
		nil,
		subscription.Data{
			Metadata: map[string]string{
				"description": "pre existing",
			},
			Route: subscription.Route{
				Destinations: []string{
					"route 1",
				},
				Condition: condition.NewKiwiTreeCondition(
					condition.NewKiwiCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
						false,
						"pattern0",
					),
				),
			},
		},
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		sd  subscription.Data
		err error
	}{
		"subscription 0": {
			err: ErrNotFound,
		},
		id1: {
			sd: subscription.Data{
				Metadata: map[string]string{
					"description": "pre existing",
				},
				Route: subscription.Route{
					Destinations: []string{
						"route 1",
					},
					Condition: condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
							false,
							"pattern0",
						),
					),
				},
			},
		},
	}
	//
	for id, c := range cases {
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			sd, err := svc.Read(ctx, id)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd.Metadata, sd.Metadata)
				assert.Equal(t, c.sd.Route.Destinations, sd.Route.Destinations)
				assert.True(t, conditionsDataEqual(c.sd.Route.Condition, sd.Route.Condition))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	id1, err := svc.Create(
		nil,
		subscription.Data{
			Metadata: map[string]string{
				"description": "pre-existing",
			},
			Route: subscription.Route{
				Destinations: []string{
					"route 1",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiTreeCondition(
							condition.NewKiwiCondition(
								condition.NewKeyCondition(
									condition.NewCondition(false),
									"", "key0",
								),
								false,
								"pattern0",
							),
						),
						condition.NewKiwiTreeCondition(
							condition.NewKiwiCondition(
								condition.NewKeyCondition(
									condition.NewCondition(true),
									"", "key1",
								),
								true,
								"pattern1",
							),
						),
					},
				),
			},
		},
	)
	require.Nil(t, err)
	id2, err := svc.Create(
		nil,
		subscription.Data{
			Metadata: map[string]string{
				"description": "fails to clean up kiwi",
			},
			Route: subscription.Route{
				Destinations: []string{
					"route 2",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiTreeCondition(
							condition.NewKiwiCondition(
								condition.NewKeyCondition(
									condition.NewCondition(false),
									"", "key0",
								),
								false,
								"pattern0",
							),
						),
						condition.NewKiwiTreeCondition(
							condition.NewKiwiCondition(
								condition.NewKeyCondition(
									condition.NewCondition(true),
									"", "key1",
								),
								true,
								"fail",
							),
						),
					},
				),
			},
		},
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id     string
		err    error
		errMsg string
	}{
		"subscription 0": {
			id:     "missing",
			err:    ErrNotFound,
			errMsg: "subscription was not found: subscription was not found by id: subscription 0",
		},
		id1: {},
		id2: {
			err:    ErrCleanKiwis,
			errMsg: fmt.Sprintf("kiwis cleanup failure, may cause kiwis garbage: internal failure, subscription id: %s", id2),
		},
	}
	//
	for id, c := range cases {
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Delete(ctx, id)
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
				assert.Equal(t, c.errMsg, err.Error())
			}
		})
	}
}

func TestService_SearchByKiwi(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := subscription.Data{
			Metadata: map[string]string{},
			Route: subscription.Route{
				Destinations: []string{
					fmt.Sprintf("route %d", i),
				},
				Condition: condition.NewKiwiTreeCondition(
					condition.NewKiwiCondition(
						condition.NewKeyCondition(
							condition.NewCondition(false), "", fmt.Sprintf("key%d", i%4),
						),
						i%3 == 2,
						fmt.Sprintf("pattern%d", i%5),
					),
				),
			},
		}
		_, err := svc.Create(ctx, req)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		query    condition.Query
		pageSize int
		err      error
	}{
		"key0/pattern0 -> 5 subs": {
			query: condition.Query{
				Limit: 10,
				Condition: condition.NewKiwiCondition(
					condition.NewKeyCondition(
						condition.NewCondition(false),
						"", "key0",
					),
					false,
					"pattern0",
				),
			},
			pageSize: 5,
		},
		"key1/pattern1, limit=2": {
			query: condition.Query{
				Limit: 2,
				Condition: condition.NewKiwiCondition(
					condition.NewKeyCondition(
						condition.NewCondition(false),
						"", "key1",
					),
					false,
					"pattern1",
				),
			},
			pageSize: 2,
		},
		"fail on group condition query": {
			query: condition.Query{
				Limit: 3,
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{},
				),
			},
			err: ErrInvalidQuery,
		},
		"fail on base condition query": {
			query: condition.Query{
				Limit:     3,
				Condition: condition.NewCondition(false),
			},
			err: ErrInvalidQuery,
		},
		"fail on key condition query": {
			query: condition.Query{
				Limit: 3,
				Condition: condition.NewKeyCondition(
					condition.NewCondition(false),
					"", "key0",
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
			page, err := svc.SearchByCondition(ctx, c.query, "")
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.pageSize, len(page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_SearchByMetadata(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := subscription.Data{
			Metadata: map[string]string{
				fmt.Sprintf("key%d", i%5): fmt.Sprintf("value%d", i%7),
			},
			Route: subscription.Route{
				Destinations: []string{
					"route0",
				},
				Condition: condition.NewKiwiTreeCondition(
					condition.NewKiwiCondition(
						condition.NewKeyCondition(
							condition.NewCondition(false), "", fmt.Sprintf("key%d", i%4),
						),
						i%3 == 2,
						fmt.Sprintf("pattern%d", i%5),
					),
				),
			},
		}
		_, err := svc.Create(ctx, req)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		query    model.MetadataQuery
		pageSize int
		err      error
	}{
		"key0/value0 -> 3 subs": {
			query: model.MetadataQuery{
				Limit: 10,
				Metadata: map[string]string{
					"key0": "value0",
				},
			},
			pageSize: 3,
		},
		"key1/value2 -> 3 subs": {
			query: model.MetadataQuery{
				Limit: 10,
				Metadata: map[string]string{
					"key1": "value2",
				},
			},
			pageSize: 3,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			page, err := svc.SearchByMetadata(ctx, c.query, "")
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.pageSize, len(page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func conditionsDataEqual(a, b condition.Condition) (equal bool) {
	equal = a.IsNot() == b.IsNot()
	if equal {
		switch at := a.(type) {
		case condition.GroupCondition:
			switch bt := b.(type) {
			case condition.GroupCondition:
				equal = at.GetLogic() == bt.GetLogic()
				if equal {
					ag := at.GetGroup()
					bg := bt.GetGroup()
					equal = len(ag) == len(bg)
					if equal {
						for i, ac := range ag {
							equal = conditionsDataEqual(ac, bg[i])
							if !equal {
								break
							}
						}
					}
				}
			case condition.KiwiCondition:
				equal = false
			default:
				equal = false
			}
		case condition.KiwiCondition:
			switch bt := b.(type) {
			case condition.GroupCondition:
				equal = false
			case condition.KiwiCondition:
				equal = at.IsPartial() == bt.IsPartial() && at.GetKey() == bt.GetKey() && at.GetPattern() == bt.GetPattern()
			default:
				equal = false
			}
		}
	}
	return equal
}
