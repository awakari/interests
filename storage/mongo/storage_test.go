package mongo

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"os"
	"testing"
	"time"
)

var (
	dbUri = os.Getenv("DB_URI_TEST_MONGO")
)

func TestNewStorage(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", rand.Uint32())
	dbCfg := config.Db{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	assert.NotNil(t, s)
	assert.Nil(t, err)
	//
	clear(ctx, t, s.(storageImpl))
}

func clear(ctx context.Context, t *testing.T, s storageImpl) {
	require.Nil(t, s.coll.Drop(ctx))
	require.Nil(t, s.Close())
}

func TestStorageImpl_Create(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", rand.Uint32())
	dbCfg := config.Db{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	id, err := s.Create(ctx, model.SubscriptionData{
		Metadata: map[string]string{
			"description": "test subscription 0",
		},
		Routes: []string{
			"test route 0",
		},
		Condition: model.NewKiwiCondition(
			model.NewKeyCondition(
				model.NewCondition(false),
				"key0",
			),
			true,
			"pattern0",
		),
	})
	assert.Nil(t, err)
	_, err = uuid.Parse(id)
	assert.Nil(t, err)
	//
	cases := map[string]struct {
		sd  model.SubscriptionData
		err error
	}{
		"success": {
			sd: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "test subscription 1",
				},
				Routes: []string{
					"test route 0",
				},
				Condition: model.NewGroupCondition(
					model.NewCondition(false),
					model.GroupLogicOr,
					[]model.Condition{
						model.NewKiwiTreeCondition(
							model.NewKiwiCondition(
								model.NewKeyCondition(
									model.NewCondition(true),
									"key0",
								),
								true,
								"pattern0",
							),
						),
						model.NewKiwiTreeCondition(
							model.NewKiwiCondition(
								model.NewKeyCondition(
									model.NewCondition(false),
									"key1",
								),
								false,
								"pattern1",
							),
						),
					},
				),
			},
		},
		"index allows duplicate kiwi in the subscription": {
			sd: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "test subscription 2",
				},
				Condition: model.NewGroupCondition(
					model.NewCondition(false),
					model.GroupLogicAnd,
					[]model.Condition{
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								"key0",
							),
							false,
							"pattern0",
						),
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								"key0",
							),
							false,
							"pattern0",
						),
					},
				),
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			id, err = s.Create(ctx, c.sd)
			if c.err == nil {
				assert.Nil(t, err)
				_, err = uuid.Parse(id)
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Read(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", rand.Uint32())
	dbCfg := config.Db{
		Uri:  dbUri,
		Name: "subscriptions-dev",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	cond0 := model.NewKiwiCondition(
		model.NewKeyCondition(
			model.NewCondition(false),
			"key0",
		),
		true,
		"pattern0",
	)
	id0, err := s.Create(ctx, model.SubscriptionData{
		Metadata: map[string]string{
			"description": "test subscription 0",
		},
		Routes: []string{
			"test route 0",
		},
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		name string
		sd   model.SubscriptionData
		err  error
	}{
		"success": {
			name: id0,
			sd: model.SubscriptionData{
				Metadata: map[string]string{
					"description": "test subscription 0",
				},
				Routes: []string{
					"test route 0",
				},
				Condition: cond0,
			},
		},
		"not found": {
			name: "sub1",
			err:  storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sub, err := s.Read(ctx, c.name)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd, sub)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Delete(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", rand.Uint32())
	dbCfg := config.Db{
		Uri:  dbUri,
		Name: "subscriptions-dev",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	cond0 := model.NewKiwiCondition(
		model.NewKeyCondition(
			model.NewCondition(false),
			"key0",
		),
		true,
		"pattern0",
	)
	id0, err := s.Create(ctx, model.SubscriptionData{
		Metadata: map[string]string{},
		Routes: []string{
			"test route 0",
		},
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		name string
		sd   model.SubscriptionData
		err  error
	}{
		"success": {
			name: id0,
			sd: model.SubscriptionData{
				Metadata: map[string]string{},
				Routes: []string{
					"test route 0",
				},
				Condition: cond0,
			},
		},
		"not found": {
			name: "sub1",
			err:  storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sub, err := s.Delete(ctx, c.name)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd, sub)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchByKiwi(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", rand.Uint32())
	dbCfg := config.Db{
		Uri:  dbUri,
		Name: "subscriptions-dev",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	var rootConditions []model.Condition
	var ids []string
	for i := 0; i < 10; i++ {
		cond := model.NewKiwiCondition(
			model.NewKeyCondition(
				model.NewCondition(i%4 == 0),
				fmt.Sprintf("key%d", i%3),
			),
			i%2 == 0,
			fmt.Sprintf("pattern%d", i%3),
		)
		sub := model.SubscriptionData{
			Metadata: map[string]string{},
			Routes: []string{
				fmt.Sprintf("test route %d", i),
			},
			Condition: cond,
		}
		id, err := s.Create(ctx, sub)
		require.Nil(t, err)
		rootConditions = append(rootConditions, cond)
		ids = append(ids, id)
	}
	//
	cases := map[string]struct {
		q      storage.KiwiQuery
		cursor string
		page   []model.Subscription
		err    error
	}{
		"1": {
			q: storage.KiwiQuery{
				Limit:   100,
				Key:     "key1",
				Pattern: "pattern1",
				Partial: true,
			},
			page: []model.Subscription{
				{
					Id: ids[4],
					Data: model.SubscriptionData{
						Metadata: map[string]string{},
						Routes: []string{
							"test route 4",
						},
						Condition: rootConditions[4],
					},
				},
			},
		},
		"2": {
			q: storage.KiwiQuery{
				Limit:   100,
				Key:     "key0",
				Pattern: "pattern0",
				Partial: false,
			},
			page: []model.Subscription{
				{
					Id: ids[3],
					Data: model.SubscriptionData{
						Metadata: map[string]string{},
						Routes: []string{
							"test route 3",
						},
						Condition: rootConditions[3],
					},
				},
				{
					Id: ids[9],
					Data: model.SubscriptionData{
						Metadata: map[string]string{},
						Routes: []string{
							"test route 9",
						},
						Condition: rootConditions[9],
					},
				},
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.SearchByKiwi(ctx, c.q, c.cursor)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, len(c.page), len(p))
				assert.ElementsMatch(t, c.page, p)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchByMetadata(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", rand.Uint32())
	dbCfg := config.Db{
		Uri:  dbUri,
		Name: "subscriptions-dev",
	}
	dbCfg.Table.Name = collName
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	var rootConditions []model.Condition
	var ids []string
	for i := 0; i < 10; i++ {
		cond := model.NewKiwiCondition(
			model.NewKeyCondition(
				model.NewCondition(i%4 == 0),
				fmt.Sprintf("key%d", i%3),
			),
			i%2 == 0,
			fmt.Sprintf("pattern%d", i%3),
		)
		sub := model.SubscriptionData{
			Metadata: map[string]string{
				fmt.Sprintf("key%d", i%2): fmt.Sprintf("value%d", i%3),
			},
			Routes: []string{
				fmt.Sprintf("test route %d", i),
			},
			Condition: cond,
		}
		id, err := s.Create(ctx, sub)
		require.Nil(t, err)
		rootConditions = append(rootConditions, cond)
		ids = append(ids, id)
	}
	//
	cases := map[string]struct {
		q        model.MetadataQuery
		cursor   string
		pageSize int
		err      error
	}{
		"1": {
			q: model.MetadataQuery{
				Limit: 100,
				Metadata: map[string]string{
					"key0": "value1",
				},
			},
			pageSize: 1,
		},
		"0": {
			q: model.MetadataQuery{
				Limit: 100,
				Metadata: map[string]string{
					"key1": "value3",
				},
			},
			pageSize: 0,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.SearchByMetadata(ctx, c.q, c.cursor)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.pageSize, len(p))
				for _, sub := range p {
					assert.Equal(t, c.q.Metadata, sub.Data.Metadata)
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
