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
	storMem := make(map[string]model.SubscriptionData)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	_, err := svc.Create(
		nil,
		model.SubscriptionData{
			Metadata: map[string]string{
				"description": "pre-existing",
			},
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
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		req model.SubscriptionData
		err error
	}{
		"empty": {
			err: model.ErrInvalidSubscription,
		},
		"locked": {
			req: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "my subscription",
				},
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
			req: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "my subscription",
				},
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
			req: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "my subscription",
				},
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
			req: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "conflict",
				},
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
	storMem := make(map[string]model.SubscriptionData)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	id1, err := svc.Create(
		nil,
		model.SubscriptionData{
			Metadata: map[string]string{
				"description": "pre existing",
			},
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
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		sd  model.SubscriptionData
		err error
	}{
		"subscription 0": {
			err: ErrNotFound,
		},
		id1: {
			sd: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "pre existing",
				},
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
	for id, c := range cases {
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			sd, err := svc.Read(ctx, id)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd.Metadata, sd.Metadata)
				assert.Equal(t, c.sd.Routes, sd.Routes)
				assert.True(t, conditionsDataEqual(c.sd.Condition, sd.Condition))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	//
	storMem := make(map[string]model.SubscriptionData)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	id1, err := svc.Create(
		nil,
		model.SubscriptionData{
			Metadata: map[string]string{
				"description": "pre-existing",
			},
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
	)
	require.Nil(t, err)
	id2, err := svc.Create(
		nil,
		model.SubscriptionData{
			Metadata: map[string]string{
				"description": "fails to clean up kiwi",
			},
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
	storMem := make(map[string]model.SubscriptionData)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := model.SubscriptionData{
			Metadata: map[string]string{},
			Routes: []string{
				fmt.Sprintf("route %d", i),
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
		_, err := svc.Create(ctx, req)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		query    model.ConditionQuery
		pageSize int
		err      error
	}{
		"key0/pattern0 -> 5 subs": {
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
			pageSize: 5,
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
			pageSize: 2,
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
	storMem := make(map[string]model.SubscriptionData)
	stor := storage.NewStorageMock(storMem)
	kiwiTreeSvc := kiwiTree.NewServiceMock()
	svc := NewService(stor, kiwiTreeSvc, kiwiTreeSvc)
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := model.SubscriptionData{
			Metadata: map[string]string{
				fmt.Sprintf("key%d", i%5): fmt.Sprintf("value%d", i%7),
			},
			Routes: []string{
				"route0",
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

func conditionsDataEqual(a, b model.Condition) (equal bool) {
	equal = a.IsNot() == b.IsNot()
	if equal {
		switch at := a.(type) {
		case model.GroupCondition:
			switch bt := b.(type) {
			case model.GroupCondition:
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
			case model.KiwiCondition:
				equal = false
			default:
				equal = false
			}
		case model.KiwiCondition:
			switch bt := b.(type) {
			case model.GroupCondition:
				equal = false
			case model.KiwiCondition:
				equal = at.IsPartial() == bt.IsPartial() && at.GetKey() == bt.GetKey() && at.GetPattern() == bt.GetPattern()
			default:
				equal = false
			}
		}
	}
	return equal
}
