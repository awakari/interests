package mongo

import (
	"context"
	"fmt"
	"github.com/meandros-messaging/subscriptions/config"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/storage"
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
		Name: "subscriptions-dev",
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
		Includes: model.MatcherGroup{
			All: true,
			Matchers: []model.Matcher{
				{
					Partial: true,
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
				Excludes: model.MatcherGroup{
					Matchers: []model.Matcher{
						{
							Partial: true,
							MatcherData: model.MatcherData{
								Key: "key0",
								Pattern: model.Pattern{
									Code: []byte("pattern0"),
									Src:  "pattern0",
								},
							},
						},
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
		"duplicate name failure": {
			sub: model.Subscription{
				Name:        "sub0",
				Description: "test subscription 123",
				Routes: []string{
					"test route 0",
				},
				Excludes: model.MatcherGroup{
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
			err: storage.ErrConflict,
		},
		"unique index allows duplicate matcher in group :(": {
			sub: model.Subscription{
				Name:        "sub2",
				Description: "test subscription 2",
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
		Includes: model.MatcherGroup{
			All: true,
			Matchers: []model.Matcher{
				{
					Partial: true,
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
				Includes: model.MatcherGroup{
					All: true,
					Matchers: []model.Matcher{
						{
							Partial: true,
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
		Includes: model.MatcherGroup{
			All: true,
			Matchers: []model.Matcher{
				{
					Partial: true,
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
				Includes: model.MatcherGroup{
					All: true,
					Matchers: []model.Matcher{
						{
							Partial: true,
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
	for i := 0; i < 4; i++ {
		err = s.Create(ctx, model.Subscription{
			Name: fmt.Sprintf("sub%d", i),
			Routes: []string{
				"test route 0",
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
		mg := model.MatcherGroup{
			Matchers: []model.Matcher{
				{
					Partial: i%2 == 0,
					MatcherData: model.MatcherData{
						Key: fmt.Sprintf("key%d", i%3),
						Pattern: model.Pattern{
							Code: []byte(fmt.Sprintf("pattern%d", i%3)),
							Src:  fmt.Sprintf("pattern%d", i%3),
						},
					},
				},
			},
		}
		sub := model.Subscription{
			Name:        fmt.Sprintf("sub%d", i),
			Description: fmt.Sprintf("test subscription %d", i),
			Routes: []string{
				fmt.Sprintf("test route %d", i),
			},
		}
		if i%4 == 0 {
			sub.Excludes = mg
		} else {
			sub.Includes = mg
		}
		err = s.Create(ctx, sub)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		q      storage.Query
		cursor string
		page   []model.Subscription
		err    error
	}{
		"1": {
			q: storage.Query{
				InExcludes: true,
				Matcher: model.Matcher{
					Partial: true,
					MatcherData: model.MatcherData{
						Key: "key0",
						Pattern: model.Pattern{
							Code: []byte("pattern0"),
							Src:  "pattern0",
						},
					},
				},
			},
			page: []model.Subscription{
				{
					Name:        "sub0",
					Description: "test subscription 0",
					Routes: []string{
						"test route 0",
					},
					Excludes: model.MatcherGroup{
						Matchers: []model.Matcher{
							{
								Partial: true,
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
		"2": {
			q: storage.Query{
				Matcher: model.Matcher{
					Partial: false,
					MatcherData: model.MatcherData{
						Key: "key0",
						Pattern: model.Pattern{
							Code: []byte("pattern0"),
							Src:  "pattern0",
						},
					},
				},
			},
			page: []model.Subscription{
				{
					Name:        "sub3",
					Description: "test subscription 3",
					Routes: []string{
						"test route 3",
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
					Name:        "sub9",
					Description: "test subscription 9",
					Routes: []string{
						"test route 9",
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
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			p, err := s.Search(ctx, c.q, c.cursor)
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
