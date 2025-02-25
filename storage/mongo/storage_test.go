package mongo

import (
	"context"
	"fmt"
	"github.com/awakari/interests/config"
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/model/interest"
	"github.com/awakari/interests/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"sort"
	"testing"
	"time"
)

var dbUri = os.Getenv("DB_URI_TEST_MONGO")

func TestNewStorage(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	err = s.Create(ctx, "interest0", "group0", "acc0", interest.Data{
		Description: "test interest 0",
		Condition: condition.NewTextCondition(
			condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
			"pattern0", false,
		),
	})
	assert.Nil(t, err)
	//
	cases := map[string]struct {
		id  string
		sd  interest.Data
		err error
	}{
		"success": {
			id: "interest1",
			sd: interest.Data{
				Description:    "test interest 1",
				Expires:        time.Now().Add(1 * time.Hour),
				Public:         true,
				LimitPerMinute: 10,
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
		"conflict": {
			id: "interest0",
			sd: interest.Data{
				Description: "test interest 1",
				Expires:     time.Now().Add(1 * time.Hour),
				Public:      true,
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
			err: storage.ErrConflict,
		},
		"index allows duplicate condition in the interest": {
			id: "interest2",
			sd: interest.Data{
				Description: "test interest 2",
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
			err = s.Create(ctx, c.id, "group0", "acc0", c.sd)
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
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	err = s.Create(ctx, "interest0", "group0", "user0", interest.Data{
		Description:    "test interest 0",
		Enabled:        true,
		Expires:        time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition:      cond0,
		Created:        time.Date(2023, 10, 4, 6, 44, 57, 0, time.UTC),
		Updated:        time.Date(2023, 10, 4, 6, 44, 58, 0, time.UTC),
		LimitPerMinute: 3,
	})
	require.Nil(t, err)
	err = s.Create(ctx, "interest1", "group1", "user1", interest.Data{
		Description: "test interest 1",
		Enabled:     true,
		Expires:     time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition:   cond0,
		Created:     time.Date(2023, 10, 4, 6, 44, 57, 0, time.UTC),
		Updated:     time.Date(2023, 10, 4, 6, 44, 58, 0, time.UTC),
		Public:      true,
		Followers:   42,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id           string
		groupId      string
		userId       string
		internal     bool
		sd           interest.Data
		ownerGroupId string
		ownerUserId  string
		err          error
	}{
		"success": {
			id:      "interest0",
			groupId: "group0",
			userId:  "user0",
			sd: interest.Data{
				Description:    "test interest 0",
				Enabled:        true,
				Expires:        time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
				Condition:      cond0,
				Created:        time.Date(2023, 10, 4, 6, 44, 57, 0, time.UTC),
				Updated:        time.Date(2023, 10, 4, 6, 44, 58, 0, time.UTC),
				LimitPerMinute: 3,
			},
			ownerGroupId: "group0",
			ownerUserId:  "user0",
		},
		"internal": {
			id:       "interest0",
			internal: true,
			sd: interest.Data{
				Description: "test interest 0",
				Enabled:     true,
				Expires:     time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
				Condition:   cond0,
				Created:     time.Date(2023, 10, 4, 6, 44, 57, 0, time.UTC),
				Updated:     time.Date(2023, 10, 4, 6, 44, 58, 0, time.UTC),
			},
			ownerGroupId: "group0",
			ownerUserId:  "user0",
		},
		"not found by id": {
			id:      "interest2",
			groupId: "group0",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
		"not found by group": {
			id:      "interest0",
			groupId: "group1",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
		"not found by user": {
			id:      "interest0",
			groupId: "group0",
			userId:  "user1",
			err:     storage.ErrNotFound,
		},
		"found public": {
			id:      "interest1",
			groupId: "group0",
			userId:  "user0",
			sd: interest.Data{
				Description: "test interest 1",
				Enabled:     true,
				Expires:     time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
				Condition:   cond0,
				Created:     time.Date(2023, 10, 4, 6, 44, 57, 0, time.UTC),
				Updated:     time.Date(2023, 10, 4, 6, 44, 58, 0, time.UTC),
				Public:      true,
				Followers:   42,
			},
			ownerGroupId: "group1",
			ownerUserId:  "user1",
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sd, ownerGroupId, ownerUserId, err := s.Read(ctx, c.id, c.groupId, c.userId, c.internal)
			if c.err == nil {
				assert.Nil(t, err)
				assert.True(t, c.sd.Condition.Equal(sd.Condition))
				assert.Equal(t, c.sd.Description, sd.Description)
				assert.Equal(t, c.sd.Enabled, sd.Enabled)
				assert.Equal(t, c.sd.Public, sd.Public)
				assert.Equal(t, c.sd.Followers, sd.Followers)
				assert.Equal(t, c.ownerGroupId, ownerGroupId)
				assert.Equal(t, c.ownerUserId, ownerUserId)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Update(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	sd0 := interest.Data{
		Expires:   time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition: cond0,
	}
	err = s.Create(ctx, "interest0", "group0", "user0", sd0)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id      string
		groupId string
		userId  string
		err     error
		sd      interest.Data
		prev    interest.Data
	}{
		"ok": {
			id:      "interest0",
			groupId: "group0",
			userId:  "user0",
			sd: interest.Data{
				Description: "new description",
				Expires:     time.Date(2023, 10, 5, 6, 44, 55, 0, time.UTC),
				Condition: condition.NewTextCondition(
					condition.NewKeyCondition(condition.NewCondition(false), "cond1", "key1"),
					"pattern1", true,
				),
				Public:         true,
				LimitPerMinute: 1,
			},
			prev: sd0,
		},
		"id mismatch": {
			id:      "interest1",
			groupId: "group0",
			userId:  "user0",
			sd: interest.Data{
				Description: "new description",
			},
			err: storage.ErrNotFound,
		},
		"acc mismatch": {
			id:      "interest0",
			groupId: "group1",
			userId:  "user0",
			sd: interest.Data{
				Description: "new description",
			},
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var prev interest.Data
			prev, err = s.Update(ctx, c.id, c.groupId, c.userId, c.sd)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.prev, prev)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Delete(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	err = s.Create(ctx, "interest0", "acc0", "user0", interest.Data{
		Expires:   time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
		Condition: cond0,
	})
	require.Nil(t, err)
	err = s.Create(ctx, "interest1", "acc0", "user1", interest.Data{
		Expires:   time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC),
		Condition: cond0,
		Public:    true,
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id      string
		groupId string
		userId  string
		sd      interest.Data
		err     error
	}{
		"success": {
			id:      "interest0",
			groupId: "acc0",
			userId:  "user0",
			sd: interest.Data{
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
			id:      "interest0",
			groupId: "acc1",
			userId:  "user0",
			err:     storage.ErrNotFound,
		},
		"cannot delete public by id": {
			id:      "interest1",
			groupId: "acc0",
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

func TestStorageImpl_Search(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
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
		sub := interest.Data{
			Description: fmt.Sprintf("description%d", i%3),
			Expires:     time.Now().Add(time.Duration(i-2) * time.Hour),
			Condition:   cond,
			Public:      i%5 == 4,
			Followers:   (10 - int64(i)) / 2,
			Created:     time.Date(2024, 2, i+1, 1, 2, 4, 5, time.UTC),
		}
		err = s.Create(ctx, fmt.Sprintf("interest%d", i), fmt.Sprintf("acc%d", i%2), fmt.Sprintf("user%d", i%2), sub)
		require.Nil(t, err)
		rootConditions = append(rootConditions, cond)
		ids = append(ids, fmt.Sprintf("interest%d", i))
	}
	fmt.Println(ids)
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
	publicIds0 := []string{
		ids[0],
		ids[2],
		ids[4],
		ids[6],
		ids[8],
		ids[9],
	}
	sort.Strings(publicIds0)
	descFollowersIds0 := []string{
		ids[0],
		ids[2],
		ids[4],
		ids[6],
	}
	descFollowersIds1 := []string{
		ids[4],
		ids[6],
		ids[8],
		ids[9],
	}
	//
	cases := map[string]struct {
		q      interest.Query
		cursor interest.Cursor
		ids    []string
		err    error
	}{
		"acc0": {
			q: interest.Query{
				Limit:   100,
				GroupId: "acc0",
				UserId:  "user0",
			},
			ids: acc0Ids,
		},
		"pattern filter": {
			q: interest.Query{
				Limit:   100,
				GroupId: "acc0",
				UserId:  "user0",
				Pattern: "description1",
			},
			ids: []string{
				ids[4],
			},
		},
		"desc": {
			q: interest.Query{
				Limit:   2,
				GroupId: "acc0",
				UserId:  "user0",
				Order:   interest.OrderDesc,
			},
			cursor: interest.Cursor{
				Id: acc0Ids[3],
			},
			ids: []string{
				acc0Ids[2],
				acc0Ids[1],
			},
		},
		"acc1": {
			q: interest.Query{
				Limit:   3,
				GroupId: "acc1",
				UserId:  "user1",
			},
			ids: acc1Ids[:3],
		},
		"include public": {
			q: interest.Query{
				Limit:         100,
				GroupId:       "acc0",
				UserId:        "user0",
				IncludePublic: true,
			},
			ids: publicIds0,
		},
		"include public and sort by followers": {
			q: interest.Query{
				Limit:         4,
				GroupId:       "acc0",
				UserId:        "user0",
				IncludePublic: true,
				Sort:          interest.SortFollowers,
				Order:         interest.OrderDesc,
			},
			cursor: interest.Cursor{
				Followers: math.MaxInt64,
			},
			ids: descFollowersIds0,
		},
		"include public and sort by followers desc w/ cursor": {
			q: interest.Query{
				Limit:         4,
				GroupId:       "acc0",
				UserId:        "user0",
				IncludePublic: true,
				Sort:          interest.SortFollowers,
				Order:         interest.OrderDesc,
			},
			cursor: interest.Cursor{
				Id:        acc0Ids[4],
				Followers: 3,
			},
			ids: descFollowersIds1,
		},
		"private only": {
			q: interest.Query{
				Limit:       100,
				GroupId:     "acc0",
				UserId:      "user0",
				PrivateOnly: true,
			},
			ids: []string{
				ids[0],
				ids[2],
				ids[6],
				ids[8],
			},
		},
		"include public and sort by creation time desc": {
			q: interest.Query{
				Limit:         3,
				GroupId:       "acc0",
				UserId:        "user0",
				IncludePublic: true,
				Sort:          interest.SortTimeCreated,
				Order:         interest.OrderDesc,
			},
			cursor: interest.Cursor{
				CreatedAt: time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC),
			},
			ids: []string{
				ids[9],
				ids[8],
				ids[6],
			},
		},
		"include public and sort by creation time asc": {
			q: interest.Query{
				Limit:         10,
				GroupId:       "acc1",
				UserId:        "user1",
				IncludePublic: true,
				Sort:          interest.SortTimeCreated,
				Order:         interest.OrderAsc,
			},
			ids: []string{
				ids[1],
				ids[3],
				ids[4],
				ids[5],
				ids[7],
				ids[9],
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.Search(ctx, c.q, c.cursor)
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
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
		sub := interest.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		id := fmt.Sprintf("interest%d", i)
		err = s.Create(ctx, id, "acc0", "user0", sub)
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
		sub := interest.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		err = s.Create(ctx, fmt.Sprintf("interest-%d", i), "acc0", "user0", sub)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		q      interest.QueryByCondition
		cursor string
		out    []interest.ConditionMatch
		err    error
	}{
		"limit=1": {
			q: interest.QueryByCondition{
				CondId: "cond0",
				Limit:  1,
			},
			out: []interest.ConditionMatch{
				{
					InterestId: matchingSubIds[0],
					Condition:  rootConditions[matchingSubIds[0]],
				},
			},
		},
		"limit=10": {
			q: interest.QueryByCondition{
				CondId: "cond0",
				Limit:  10,
			},
			out: []interest.ConditionMatch{
				{
					InterestId: matchingSubIds[0],
					Condition:  rootConditions[matchingSubIds[0]],
				},
				{
					InterestId: matchingSubIds[1],
					Condition:  rootConditions[matchingSubIds[1]],
				},
				{
					InterestId: matchingSubIds[2],
					Condition:  rootConditions[matchingSubIds[2]],
				},
				{
					InterestId: matchingSubIds[3],
					Condition:  rootConditions[matchingSubIds[3]],
				},
				{
					InterestId: matchingSubIds[4],
					Condition:  rootConditions[matchingSubIds[4]],
				},
			},
		},
		"with cursor": {
			q: interest.QueryByCondition{
				CondId: "cond0",
				Limit:  10,
			},
			cursor: matchingSubIds[2],
			out: []interest.ConditionMatch{
				{
					InterestId: matchingSubIds[3],
					Condition:  rootConditions[matchingSubIds[3]],
				},
				{
					InterestId: matchingSubIds[4],
					Condition:  rootConditions[matchingSubIds[4]],
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
					assert.Equal(t, cm.InterestId, page[i].InterestId)
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
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	sub0 := interest.Data{
		Enabled:   true,
		Condition: cond0,
	}
	err = s.Create(ctx, "interest0", "acc0", "user0", sub0)
	require.Nil(t, err)
	// already expired
	sub1 := interest.Data{
		Enabled:   true,
		Condition: cond0,
		Expires:   time.Date(2022, 2, 22, 22, 22, 22, 0, time.UTC),
	}
	err = s.Create(ctx, "interest1", "acc0", "user0", sub1)
	require.Nil(t, err)
	// not expired
	sub2 := interest.Data{
		Enabled:   true,
		Condition: cond0,
		Expires:   time.Now().Add(1 * time.Hour).UTC(),
	}
	err = s.Create(ctx, "interest2", "acc0", "user0", sub2)
	require.Nil(t, err)
	// temporarily disabled in past but active now
	sub3 := interest.Data{
		Enabled:   true,
		Condition: cond0,
	}
	err = s.Create(ctx, "interest3", "acc0", "user0", sub3)
	require.Nil(t, err)
	_, err = s.SetEnabledBatch(ctx, []string{"interest3"}, true, time.Date(2022, 2, 22, 22, 22, 22, 0, time.UTC))
	require.Nil(t, err)
	// not yet
	sub4 := interest.Data{
		Enabled:   true,
		Condition: cond0,
	}
	err = s.Create(ctx, "interest4", "acc0", "user0", sub4)
	require.Nil(t, err)
	_, err = s.SetEnabledBatch(ctx, []string{"interest4"}, true, time.Now().Add(1*time.Hour).UTC())
	require.Nil(t, err)
	//
	cases := map[string]struct {
		q      interest.QueryByCondition
		cursor string
		out    []interest.ConditionMatch
		err    error
	}{
		"2": {
			q: interest.QueryByCondition{
				CondId: "cond0",
				Limit:  10,
			},
			out: []interest.ConditionMatch{
				{
					InterestId: "interest0",
					Condition:  cond0,
				},
				{
					InterestId: "interest2",
					Condition:  cond0,
				},
				{
					InterestId: "interest3",
					Condition:  cond0,
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
				var found bool
				for _, cm := range c.out {
					for _, actual := range page {
						if cm.InterestId == actual.InterestId && cm.Condition.Equal(actual.Condition) {
							found = true
							break
						}
					}
					assert.True(t, found)
				}
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_Count(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
		sub := interest.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		id := fmt.Sprintf("interest%d", i)
		err = s.Create(ctx, id, "acc0", "user0", sub)
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
		sub := interest.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		err = s.Create(ctx, fmt.Sprintf("interest-%d", i), "acc0", "user0", sub)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		out int64
		err error
	}{
		"20": {
			out: 20,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			out, err := s.Count(context.TODO())
			assert.Equal(t, c.out, out)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageImpl_CountUsersUnique(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
		sub := interest.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		id := fmt.Sprintf("interest%d", i)
		err = s.Create(ctx, id, "acc0", fmt.Sprintf("user%d", i%3), sub)
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
		sub := interest.Data{
			Enabled:   i%2 == 0,
			Condition: cond,
		}
		err = s.Create(ctx, fmt.Sprintf("interest-%d", i), "acc0", "user0", sub)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		out int64
		err error
	}{
		"3": {
			out: 3,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			out, err := s.(storageImpl).CountUsersUnique(context.TODO())
			assert.Equal(t, c.out, out)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageImpl_UpdateFollowers(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	sd0 := interest.Data{
		Expires:   time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition: cond0,
		Public:    true,
	}
	err = s.Create(ctx, "interest0", "group0", "user0", sd0)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id       string
		newCount int64
		err      error
	}{
		"ok0": {
			id:       "interest0",
			newCount: 0,
		},
		"ok1": {
			id:       "interest0",
			newCount: 1,
		},
		"ok2": {
			id:       "interest0",
			newCount: 2,
		},
		"id mismatch": {
			id:  "missing",
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err = s.UpdateFollowers(ctx, c.id, c.newCount)
			if c.err == nil {
				assert.Nil(t, err)
				var sd interest.Data
				sd, _, _, err = s.Read(ctx, c.id, "group0", "user0", false)
				require.Nil(t, err)
				assert.Equal(t, c.newCount, sd.Followers)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_UpdateResultTime(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	sd0 := interest.Data{
		Expires:   time.Date(2023, 10, 4, 6, 44, 55, 0, time.UTC),
		Condition: cond0,
		Public:    true,
	}
	err = s.Create(ctx, "interest0", "group0", "user0", sd0)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id   string
		last time.Time
		err  error
	}{
		"ok0": {
			id:   "interest0",
			last: time.Date(2024, 8, 12, 18, 7, 0, 0, time.UTC),
		},
		"id mismatch": {
			id:  "missing",
			err: storage.ErrNotFound,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err = s.UpdateResultTime(ctx, c.id, c.last)
			if c.err == nil {
				assert.Nil(t, err)
				var sd interest.Data
				sd, _, _, err = s.Read(ctx, c.id, "group0", "user0", false)
				require.Nil(t, err)
				assert.Equal(t, c.last, sd.Result)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestStorageImpl_SetEnabledBatch(t *testing.T) {
	//
	collName := fmt.Sprintf("interests-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "interests",
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
	sd0 := interest.Data{
		Condition: cond0,
		Enabled:   true,
	}
	err = s.Create(ctx, "interest0", "group0", "user0", sd0)
	require.Nil(t, err)
	err = s.Create(ctx, "interest1", "group0", "user0", sd0)
	require.Nil(t, err)
	sd1 := interest.Data{
		Condition: cond0,
		Enabled:   false,
	}
	err = s.Create(ctx, "interest2", "group0", "user0", sd1)
	require.Nil(t, err)
	err = s.Create(ctx, "interest3", "group0", "user0", sd1)
	require.Nil(t, err)
	err = s.Create(ctx, "interest4", "group0", "user0", sd1)
	require.Nil(t, err)
	err = s.Create(ctx, "interest5", "group0", "user0", sd1)
	require.Nil(t, err)
	err = s.Create(ctx, "interest7", "group0", "user0", sd0)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		ids          []string
		enabled      bool
		enabledSince time.Time
		n            int64
		err          error
	}{
		"disable": {
			ids: []string{"interest0", "interest1", "interest2"},
			n:   2,
		},
		"enable": {
			ids:          []string{"interest3", "interest4", "interest5"},
			enabled:      true,
			enabledSince: time.Now(),
			n:            3,
		},
		"some missing": {
			ids: []string{"interest6", "interest7"},
			n:   1,
		},
		"none": {
			ids: []string{"interest8", "interest9"},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var n int64
			n, err = s.SetEnabledBatch(context.TODO(), c.ids, c.enabled, c.enabledSince)
			assert.Equal(t, c.n, n)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
