package mongo

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
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
	err = s.Create(ctx, model.Subscription{
		Name:        "sub0",
		Description: "test subscription 0",
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
	//
	cases := map[string]struct {
		sub model.Subscription
		err error
	}{
		"success": {
			sub: model.Subscription{
				Name:        "sub1",
				Description: "test subscription 1",
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
		"duplicate name failure": {
			sub: model.Subscription{
				Name:        "sub0",
				Description: "test subscription 123",
				Routes: []string{
					"test route 0",
				},
				Condition: model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						"key0",
					),
					false,
					"pattern0",
				),
			},
			err: storage.ErrConflict,
		},
		"index allows duplicate kiwi in the subscription": {
			sub: model.Subscription{
				Name:        "sub2",
				Description: "test subscription 2",
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
			err = s.Create(ctx, c.sub)
			if c.err == nil {
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
	err = s.Create(ctx, model.Subscription{
		Name:        "sub0",
		Description: "test subscription 0",
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
	require.Nil(t, err)
	//
	cases := map[string]struct {
		name string
		sub  model.Subscription
		err  error
	}{
		"success": {
			name: "sub0",
			sub: model.Subscription{
				Name:        "sub0",
				Description: "test subscription 0",
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
				assert.Equal(t, c.sub, sub)
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
	err = s.Create(ctx, model.Subscription{
		Name:        "sub0",
		Description: "test subscription 0",
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
	require.Nil(t, err)
	//
	cases := map[string]struct {
		name string
		sub  model.Subscription
		err  error
	}{
		"success": {
			name: "sub0",
			sub: model.Subscription{
				Name:        "sub0",
				Description: "test subscription 0",
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
				assert.Equal(t, c.sub, sub)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_ListNames(t *testing.T) {
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
	for i := 0; i < 4; i++ {
		err = s.Create(ctx, model.Subscription{
			Name: fmt.Sprintf("sub%d", i),
			Routes: []string{
				"test route 0",
			},
			Condition: model.NewKiwiCondition(
				model.NewKeyCondition(
					model.NewCondition(false),
					"key0",
				),
				false,
				"pattern0",
			),
		})
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		cursor string
		limit  uint32
		page   []string
		err    error
	}{
		"all at once": {
			limit: 1000,
			page: []string{
				"sub0",
				"sub1",
				"sub2",
				"sub3",
			},
		},
		"limit = 2": {
			limit: 2,
			page: []string{
				"sub0",
				"sub1",
			},
		},
		"limit = 0": {
			limit: 0,
			page: []string{
				"sub0",
				"sub1",
				"sub2",
				"sub3",
			},
		},
		"limit + cursor": {
			limit:  2,
			cursor: "sub1",
			page: []string{
				"sub2",
				"sub3",
			},
		},
		"cursor after end": {
			limit:  1000,
			cursor: "sub3",
			page:   []string{},
		},
		"cursor before begin": {
			limit:  1000,
			cursor: "abc0",
			page: []string{
				"sub0",
				"sub1",
				"sub2",
				"sub3",
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.ListNames(ctx, c.limit, c.cursor)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, len(c.page), len(p))
				for i, name := range c.page {
					assert.Equal(t, name, p[i])
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Search(t *testing.T) {
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
	for i := 0; i < 10; i++ {
		sub := model.Subscription{
			Name:        fmt.Sprintf("sub%d", i),
			Description: fmt.Sprintf("test subscription %d", i),
			Routes: []string{
				fmt.Sprintf("test route %d", i),
			},
			Condition: model.NewKiwiCondition(
				model.NewKeyCondition(
					model.NewCondition(i%4 == 0),
					fmt.Sprintf("key%d", i%3),
				),
				i%2 == 0,
				fmt.Sprintf("pattern%d", i%3),
			),
		}
		err = s.Create(ctx, sub)
		require.Nil(t, err)
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
					Name: "sub4",
					Routes: []string{
						"test route 4",
					},
					Condition: model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(true),
							fmt.Sprintf("key1"),
						),
						true,
						fmt.Sprintf("pattern1"),
					),
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
					Name: "sub3",
					Routes: []string{
						"test route 3",
					},
					Condition: model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							fmt.Sprintf("key0"),
						),
						false,
						fmt.Sprintf("pattern0"),
					),
				},
				{
					Name: "sub9",
					Routes: []string{
						"test route 9",
					},
					Condition: model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							fmt.Sprintf("key0"),
						),
						false,
						fmt.Sprintf("pattern0"),
					),
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
				for i, sub := range c.page {
					assert.Equal(t, sub, p[i])
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
