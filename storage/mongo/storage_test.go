package mongo

import (
	"context"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sort"
	"testing"
	"time"
)

var dbUri = os.Getenv("DB_URI_TEST_MONGO")

func TestNewStorage(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	id, err := s.Create(ctx, "acc0", subscription.Data{
		Metadata: subscription.Metadata{
			Description: "test subscription 0",
		},
		Condition: condition.NewKiwiCondition(
			condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
			true,
			"pattern0",
		),
	})
	assert.Nil(t, err)
	_, err = uuid.Parse(id)
	assert.Nil(t, err)
	//
	cases := map[string]struct {
		sd  subscription.Data
		err error
	}{
		"success": {
			sd: subscription.Data{
				Metadata: subscription.Metadata{
					Description: "test subscription 1",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicOr,
					[]condition.Condition{
						condition.NewKiwiTreeCondition(
							condition.NewKiwiCondition(
								condition.NewKeyCondition(
									condition.NewCondition(true), "cond0", "key0",
								),
								true,
								"pattern0",
							),
						),
						condition.NewKiwiTreeCondition(
							condition.NewKiwiCondition(
								condition.NewKeyCondition(
									condition.NewCondition(false), "cond1", "key1",
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
			sd: subscription.Data{
				Metadata: subscription.Metadata{
					Description: "test subscription 2",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
							false,
							"pattern0",
						),
						condition.NewKiwiCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "cond1", "key0"),
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
			id, err = s.Create(ctx, "acc0", c.sd)
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
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	cond0 := condition.NewKiwiCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
		true,
		"pattern0",
	)
	id0, err := s.Create(ctx, "acc0", subscription.Data{
		Metadata: subscription.Metadata{
			Description: "test subscription 0",
		},
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id  string
		acc string
		sd  subscription.Data
		err error
	}{
		"success": {
			id:  id0,
			acc: "acc0",
			sd: subscription.Data{
				Metadata: subscription.Metadata{
					Description: "test subscription 0",
				},
				Condition: cond0,
			},
		},
		"not found by id": {
			id:  "sub1",
			acc: "acc0",
			err: storage.ErrNotFound,
		},
		"not found by acc": {
			id:  id0,
			acc: "acc1",
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sd, err := s.Read(ctx, c.id, c.acc)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd.Metadata, sd.Metadata)
				assert.True(t, c.sd.Condition.Equal(sd.Condition))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_UpdateMetadata(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	cond0 := condition.NewKiwiCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
		true,
		"pattern0",
	)
	id0, err := s.Create(ctx, "acc0", subscription.Data{
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id  string
		acc string
		err error
		md  subscription.Metadata
	}{
		"ok": {
			id:  id0,
			acc: "acc0",
			md: subscription.Metadata{
				Description: "new description",
			},
		},
		"id mismatch": {
			id:  "id0",
			acc: "acc0",
			md: subscription.Metadata{
				Description: "new description",
			},
			err: storage.ErrNotFound,
		},
		"acc mismatch": {
			id:  id0,
			acc: "acc1",
			md: subscription.Metadata{
				Description: "new description",
			},
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err = s.UpdateMetadata(ctx, c.id, c.acc, c.md)
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Delete(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	cond0 := condition.NewKiwiCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
		true,
		"pattern0",
	)
	id0, err := s.Create(ctx, "acc0", subscription.Data{
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id  string
		acc string
		sd  subscription.Data
		err error
	}{
		"success": {
			id:  id0,
			acc: "acc0",
			sd: subscription.Data{
				Condition: cond0,
			},
		},
		"not found by id": {
			id:  "sub1",
			acc: "acc0",
			err: storage.ErrNotFound,
		},
		"not found by acc": {
			id:  id0,
			acc: "acc1",
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sd, err := s.Delete(ctx, c.id, c.acc)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd.Metadata, sd.Metadata)
				assert.True(t, c.sd.Condition.Equal(sd.Condition))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchByAccount(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	var rootConditions []condition.Condition
	var ids []string
	for i := 0; i < 10; i++ {
		cond := condition.NewKiwiCondition(
			condition.NewKeyCondition(
				condition.NewCondition(i%4 == 0), fmt.Sprintf("cond%d", i), fmt.Sprintf("key%d", i%3),
			),
			i%2 == 0,
			fmt.Sprintf("pattern%d", i%3),
		)
		sub := subscription.Data{
			Metadata: subscription.Metadata{
				Description: fmt.Sprintf("description%d", i%2),
			},
			Condition: cond,
		}
		id, err := s.Create(ctx, fmt.Sprintf("acc%d", i%2), sub)
		require.Nil(t, err)
		rootConditions = append(rootConditions, cond)
		ids = append(ids, id)
	}
	acc0Ids := []string{
		ids[0],
		ids[2],
		ids[4],
		ids[6],
		ids[8],
	}
	sort.Strings(acc0Ids)
	acc1Ids := []string{
		ids[1],
		ids[3],
		ids[5],
		ids[7],
		ids[9],
	}
	sort.Strings(acc1Ids)
	//
	cases := map[string]struct {
		q      subscription.QueryByAccount
		cursor string
		ids    []string
		err    error
	}{
		"acc0": {
			q: subscription.QueryByAccount{
				Limit:   100,
				Account: "acc0",
			},
			ids: acc0Ids,
		},
		"acc1": {
			q: subscription.QueryByAccount{
				Limit:   3,
				Account: "acc1",
			},
			ids: acc1Ids[:3],
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.SearchByAccount(ctx, c.q, c.cursor)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.ids, p)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchByKiwi(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
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
	var rootConditions []condition.Condition
	var ids []string
	for i := 0; i < 10; i++ {
		cond := condition.NewKiwiCondition(
			condition.NewKeyCondition(
				condition.NewCondition(i%4 == 0), fmt.Sprintf("cond%d", i), fmt.Sprintf("key%d", i%3),
			),
			i%2 == 0,
			fmt.Sprintf("pattern%d", i%3),
		)
		sub := subscription.Data{
			Metadata: subscription.Metadata{
				Enabled: i > 0,
			},
			Condition: cond,
		}
		id, err := s.Create(ctx, "acc0", sub)
		require.Nil(t, err)
		rootConditions = append(rootConditions, cond)
		ids = append(ids, id)
	}
	//
	cases := map[string]struct {
		q   storage.KiwiQuery
		out []*subscription.ConditionMatch
		err error
	}{
		"1": {
			q: storage.KiwiQuery{
				Key:     "key1",
				Pattern: "pattern1",
				Partial: true,
			},
			out: []*subscription.ConditionMatch{
				{
					SubscriptionId: ids[4],
					Condition:      rootConditions[4],
				},
			},
		},
		"2": {
			q: storage.KiwiQuery{
				Key:     "key2",
				Pattern: "pattern2",
				Partial: true,
			},
			out: []*subscription.ConditionMatch{
				{
					SubscriptionId: ids[2],
					Condition:      rootConditions[2],
				},
				{
					SubscriptionId: ids[8],
					Condition:      rootConditions[8],
				},
			},
		},
		"skip disabled subscriptions": {
			q: storage.KiwiQuery{
				Key:     "key0",
				Pattern: "pattern0",
				Partial: true,
			},
			out: []*subscription.ConditionMatch{
				{
					SubscriptionId: ids[6],
					Condition:      rootConditions[6],
				},
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var out []*subscription.ConditionMatch
			consumer := func(cm *subscription.ConditionMatch) (err error) {
				out = append(out, cm)
				return
			}
			err = s.SearchByKiwi(ctx, c.q, consumer)
			if c.err == nil {
				require.Nil(t, err)
				require.Equal(t, len(c.out), len(out))
				for i, cm := range c.out {
					assert.Equal(t, cm.SubscriptionId, out[i].SubscriptionId)
					assert.Equal(t, out[i].ConditionId, out[i].Condition.(condition.KiwiCondition).GetId())
					assert.True(t, cm.Condition.Equal(out[i].Condition))
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
