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
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
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
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	id, err := s.Create(ctx, "group0", "acc0", subscription.Data{
		Description: "test subscription 0",
		Condition: condition.NewTextCondition(
			condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
			"pattern0", false,
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
				Description: "test subscription 1",
				Expires:     time.Now().Add(1 * time.Hour),
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicOr,
					[]condition.Condition{
						condition.NewTextCondition(
							condition.NewKeyCondition(
								condition.NewCondition(true), "cond0", "key0",
							),
							"pattern0", true,
						),
						condition.NewNumberCondition(
							condition.NewKeyCondition(
								condition.NewCondition(false), "cond1", "key1",
							),
							condition.NumOpEq, 42,
						),
					},
				),
			},
		},
		"index allows duplicate condition in the subscription": {
			sd: subscription.Data{
				Description: "test subscription 2",
				Expires:     time.Now().Add(1 * time.Hour),
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewTextCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
							"pattern0", true,
						),
						condition.NewTextCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "cond1", "key0"),
							"pattern0", false,
						),
					},
				),
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			id, err = s.Create(ctx, "group0", "acc0", c.sd)
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
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	cond0 := condition.NewGroupCondition(
		condition.NewCondition(false),
		condition.GroupLogicOr,
		[]condition.Condition{
			condition.NewTextCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
				"pattern0", true,
			),
			condition.NewNumberCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
				condition.NumOpLt, -1.2e-3,
			),
		},
	)
	id0, err := s.Create(ctx, "group0", "user0", subscription.Data{
		Description: "test subscription 0",
		Enabled:     true,
		Expires:     time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition:   cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id      string
		groupId string
		userId  string
		sd      subscription.Data
		err     error
	}{
		"success": {
			id:      id0,
			groupId: "group0",
			userId:  "user0",
			sd: subscription.Data{
				Description: "test subscription 0",
				Enabled:     true,
				Expires:     time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
				Condition:   cond0,
			},
		},
		"not found by id": {
			id:      "sub1",
			groupId: "group0",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
		"not found by group": {
			id:      id0,
			groupId: "group1",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
		"not found by user": {
			id:      id0,
			groupId: "group0",
			userId:  "user1",
			err:     storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sd, err := s.Read(ctx, c.id, c.groupId, c.userId)
			if c.err == nil {
				assert.Nil(t, err)
				assert.True(t, c.sd.Condition.Equal(sd.Condition))
				assert.Equal(t, c.sd.Description, sd.Description)
				assert.Equal(t, c.sd.Enabled, sd.Enabled)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Update(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	cond0 := condition.NewTextCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
		"pattern0", false,
	)
	id0, err := s.Create(ctx, "group0", "user0", subscription.Data{
		Expires:   time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id      string
		groupId string
		userId  string
		err     error
		sd      subscription.Data
	}{
		"ok": {
			id:      id0,
			groupId: "group0",
			userId:  "user0",
			sd: subscription.Data{
				Description: "new description",
				Expires:     time.Date(2023, 10, 5, 6, 44, 55, 0, time.UTC),
			},
		},
		"id mismatch": {
			id:      "id0",
			groupId: "group0",
			userId:  "user0",
			sd: subscription.Data{
				Description: "new description",
			},
			err: storage.ErrNotFound,
		},
		"acc mismatch": {
			id:      id0,
			groupId: "group1",
			userId:  "user0",
			sd: subscription.Data{
				Description: "new description",
			},
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err = s.Update(ctx, c.id, c.groupId, c.userId, c.sd)
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
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	cond0 := condition.NewTextCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
		"pattern0", false,
	)
	id0, err := s.Create(ctx, "acc0", "user0", subscription.Data{
		Expires:   time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
		Condition: cond0,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id      string
		groupId string
		userId  string
		sd      subscription.Data
		err     error
	}{
		"success": {
			id:      id0,
			groupId: "acc0",
			userId:  "user0",
			sd: subscription.Data{
				Expires:   time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
				Condition: cond0,
			},
		},
		"not found by id": {
			id:      "sub1",
			groupId: "acc0",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
		"not found by acc": {
			id:      id0,
			groupId: "acc1",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sd, err := s.Delete(ctx, c.id, c.groupId, c.userId)
			if c.err == nil {
				assert.Nil(t, err)
				assert.True(t, c.sd.Condition.Equal(sd.Condition))
				assert.Equal(t, c.sd.Description, sd.Description)
				assert.Equal(t, c.sd.Enabled, sd.Enabled)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchOwn(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	var rootConditions []condition.Condition
	var ids []string
	for i := 0; i < 10; i++ {
		cond := condition.NewTextCondition(
			condition.NewKeyCondition(
				condition.NewCondition(i%4 == 0), fmt.Sprintf("cond%d", i), fmt.Sprintf("key%d", i%3),
			),
			fmt.Sprintf("pattern%d", i%3), i%2 == 0,
		)
		sub := subscription.Data{
			Description: fmt.Sprintf("description%d", i%2),
			Expires:     time.Now().Add(time.Duration(i-2) * time.Hour),
			Condition:   cond,
		}
		id, err := s.Create(ctx, fmt.Sprintf("acc%d", i%2), fmt.Sprintf("user%d", i%2), sub)
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
		q      subscription.QueryOwn
		cursor string
		ids    []string
		err    error
	}{
		"acc0": {
			q: subscription.QueryOwn{
				Limit:   100,
				GroupId: "acc0",
				UserId:  "user0",
			},
			ids: acc0Ids,
		},
		"acc1": {
			q: subscription.QueryOwn{
				Limit:   3,
				GroupId: "acc1",
				UserId:  "user1",
			},
			ids: acc1Ids[:3],
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.SearchOwn(ctx, c.q, c.cursor)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.ids, p)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchByCondition(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Table.Shard = false
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	//
	rootConditions := map[string]condition.Condition{}
	var matchingSubIds []string
	for i := 0; i < 10; i++ {
		cond := condition.NewTextCondition(
			condition.NewKeyCondition(
				condition.NewCondition(i%4 == 0), "cond0", fmt.Sprintf("key%d", i%3),
			),
			fmt.Sprintf("pattern%d", i%3), i%2 == 0,
		)
		sub := subscription.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		id, err := s.Create(ctx, "acc0", "user0", sub)
		require.Nil(t, err)
		if sub.Enabled {
			matchingSubIds = append(matchingSubIds, id)
			rootConditions[id] = cond
		}
	}
	sort.Strings(matchingSubIds)
	for i := 0; i < 10; i++ {
		cond := condition.NewNumberCondition(
			condition.NewKeyCondition(
				condition.NewCondition(i%4 == 0), "cond1", fmt.Sprintf("key%d", i%3),
			),
			condition.NumOpEq, 42,
		)
		sub := subscription.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		_, err := s.Create(ctx, "acc0", "user0", sub)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		q      subscription.QueryByCondition
		cursor string
		out    []subscription.ConditionMatch
		err    error
	}{
		"limit=1": {
			q: subscription.QueryByCondition{
				CondId: "cond0",
				Limit:  1,
			},
			out: []subscription.ConditionMatch{
				{
					SubscriptionId: matchingSubIds[0],
					Condition:      rootConditions[matchingSubIds[0]],
				},
			},
		},
		"limit=10": {
			q: subscription.QueryByCondition{
				CondId: "cond0",
				Limit:  10,
			},
			out: []subscription.ConditionMatch{
				{
					SubscriptionId: matchingSubIds[0],
					Condition:      rootConditions[matchingSubIds[0]],
				},
				{
					SubscriptionId: matchingSubIds[1],
					Condition:      rootConditions[matchingSubIds[1]],
				},
				{
					SubscriptionId: matchingSubIds[2],
					Condition:      rootConditions[matchingSubIds[2]],
				},
				{
					SubscriptionId: matchingSubIds[3],
					Condition:      rootConditions[matchingSubIds[3]],
				},
				{
					SubscriptionId: matchingSubIds[4],
					Condition:      rootConditions[matchingSubIds[4]],
				},
			},
		},
		"with cursor": {
			q: subscription.QueryByCondition{
				CondId: "cond0",
				Limit:  10,
			},
			cursor: matchingSubIds[2],
			out: []subscription.ConditionMatch{
				{
					SubscriptionId: matchingSubIds[3],
					Condition:      rootConditions[matchingSubIds[3]],
				},
				{
					SubscriptionId: matchingSubIds[4],
					Condition:      rootConditions[matchingSubIds[4]],
				},
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			page, err := s.SearchByCondition(ctx, c.q, c.cursor)
			if c.err == nil {
				require.Nil(t, err)
				require.Equal(t, len(c.out), len(page))
				for i, cm := range c.out {
					assert.Equal(t, cm.SubscriptionId, page[i].SubscriptionId)
					assert.True(t, cm.Condition.Equal(page[i].Condition))
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SearchByCondition_WithExpiration(t *testing.T) {
	//
	collName := fmt.Sprintf("subscriptions-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "subscriptions",
	}
	dbCfg.Table.Name = collName
	dbCfg.Table.Shard = false
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	defer clear(ctx, t, s.(storageImpl))
	// expiration not set
	cond0 := condition.NewTextCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
		"term0",
		false,
	)
	sub0 := subscription.Data{
		Enabled:   true,
		Condition: cond0,
	}
	id0, err := s.Create(ctx, "acc0", "user0", sub0)
	require.Nil(t, err)
	// already expired
	cond1 := condition.NewTextCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key1"),
		"term1",
		false,
	)
	sub1 := subscription.Data{
		Enabled:   true,
		Condition: cond1,
		Expires:   time.Date(2022, 2, 22, 22, 22, 22, 0, time.UTC),
	}
	_, err = s.Create(ctx, "acc0", "user0", sub1)
	require.Nil(t, err)
	// not expired
	cond2 := condition.NewTextCondition(
		condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key2"),
		"term2",
		false,
	)
	sub2 := subscription.Data{
		Enabled:   true,
		Condition: cond2,
		Expires:   time.Now().Add(1 * time.Hour).UTC(),
	}
	id2, err := s.Create(ctx, "acc0", "user0", sub2)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		q      subscription.QueryByCondition
		cursor string
		out    []subscription.ConditionMatch
		err    error
	}{
		"2": {
			q: subscription.QueryByCondition{
				CondId: "cond0",
				Limit:  10,
			},
			out: []subscription.ConditionMatch{
				{
					SubscriptionId: id2,
					Condition:      cond2,
				},
				{
					SubscriptionId: id0,
					Condition:      cond0,
				},
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			page, err := s.SearchByCondition(ctx, c.q, c.cursor)
			if c.err == nil {
				require.Nil(t, err)
				require.Equal(t, len(c.out), len(page))
				for i, cm := range c.out {
					assert.Equal(t, cm.SubscriptionId, page[i].SubscriptionId)
					assert.True(t, cm.Condition.Equal(page[i].Condition))
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
