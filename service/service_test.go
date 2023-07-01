package service

import (
	"context"
	"fmt"
	conditions_text "github.com/awakari/subscriptions/api/grpc/conditions-text"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	condTextSvc := conditions_text.NewServiceMock()
	svc := NewService(stor, condTextSvc)
	svc = NewLoggingMiddleware(svc, slog.Default())
	_, err := svc.Create(
		nil,
		"acc0",
		"user0",
		subscription.Data{
			Description: "pre-existing",
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
				"pattern0", false,
			),
		},
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		req subscription.Data
		err error
	}{
		"empty": {
			err: subscription.ErrInvalidSubscriptionCondition,
		},
		"locked": {
			req: subscription.Data{
				Description: "my subscription",
				Condition: condition.NewTextCondition(
					condition.NewKeyCondition(condition.NewCondition(false), "", "conflict"),
					"locked", false,
				),
			},
			err: ErrShouldRetry,
		},
		"fail": {
			req: subscription.Data{
				Description: "my subscription",
				Condition: condition.NewTextCondition(
					condition.NewKeyCondition(
						condition.NewCondition(false),
						"", "fail",
					),
					"fail", false,
				),
			},
			err: ErrInternal,
		},
		"ok": {
			req: subscription.Data{
				Description: "my subscription",
				Condition: condition.NewTextCondition(
					condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
					"ok", false,
				),
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			id, err := svc.Create(ctx, "acc0", "user0", c.req)
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
	condTxtSvc := conditions_text.NewServiceMock()
	svc := NewService(stor, condTxtSvc)
	svc = NewLoggingMiddleware(svc, slog.Default())
	id1, err := svc.Create(
		nil,
		"acc0",
		"user0",
		subscription.Data{

			Description: "pre existing",
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
				"pattern0", false,
			),
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

				Description: "pre existing",
				Condition: condition.NewTextCondition(
					condition.NewKeyCondition(condition.NewCondition(false), "", "key0"),
					"pattern0", false,
				),
			},
		},
	}
	//
	for id, c := range cases {
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			sd, err := svc.Read(ctx, id, "acc0", "user0")
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.sd.Enabled, sd.Enabled)
				assert.Equal(t, c.sd.Description, sd.Description)
				assert.True(t, conditionsDataEqual(c.sd.Condition, sd.Condition))
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
	condTextSvc := conditions_text.NewServiceMock()
	svc := NewService(stor, condTextSvc)
	svc = NewLoggingMiddleware(svc, slog.Default())
	id1, err := svc.Create(
		nil,
		"acc0",
		"user0",
		subscription.Data{

			Description: "pre-existing",
			Condition: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicAnd,
				[]condition.Condition{
					condition.NewTextCondition(
						condition.NewKeyCondition(
							condition.NewCondition(false),
							"", "key0",
						),
						"pattern0", false,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(
							condition.NewCondition(true),
							"", "key1",
						),
						"pattern1", false,
					),
				},
			),
		},
	)
	require.Nil(t, err)
	id2, err := svc.Create(
		nil,
		"acc0",
		"user0",
		subscription.Data{
			Description: "fails to clean up conditions",
			Condition: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicAnd,
				[]condition.Condition{
					condition.NewTextCondition(
						condition.NewKeyCondition(
							condition.NewCondition(false),
							"", "key0",
						),
						"pattern0", false,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(
							condition.NewCondition(true),
							"", "fail_lock",
						),
						"pattern0", false,
					),
				},
			),
		},
	)
	require.Nil(t, err)
	//
	cases := map[string]struct {
		id  string
		err error
	}{
		"subscription 0": {
			id:  "missing",
			err: ErrNotFound,
		},
		id1: {},
		id2: {
			err: ErrCleanConditions,
		},
	}
	//
	for id, c := range cases {
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := svc.Delete(ctx, id, "acc0", "user0")
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_SearchByCondition(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	condTextSvc := conditions_text.NewServiceMock()
	svc := NewService(stor, condTextSvc)
	svc = NewLoggingMiddleware(svc, slog.Default())
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := subscription.Data{
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(
					condition.NewCondition(false), fmt.Sprintf("cond%d", i%3), fmt.Sprintf("key%d", i%4),
				),
				fmt.Sprintf("pattern%d", i%5), false,
			),
		}
		_, err := svc.Create(ctx, "acc0", "user0", req)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		condId   string
		pageSize int
		err      error
	}{
		"key0/pattern0 -> 5 subs": {
			condId:   "cond0",
			pageSize: 10000,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			var page []*subscription.ConditionMatch
			consume := func(match *subscription.ConditionMatch) (err error) {
				page = append(page, match)
				return
			}
			err := svc.SearchByCondition(ctx, c.condId, consume)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.pageSize, len(page))
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestService_SearchOwn(t *testing.T) {
	//
	storMem := make(map[string]subscription.Data)
	stor := storage.NewStorageMock(storMem)
	condTextSvc := conditions_text.NewServiceMock()
	svc := NewService(stor, condTextSvc)
	svc = NewLoggingMiddleware(svc, slog.Default())
	//
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 100; i++ {
		req := subscription.Data{

			Description: fmt.Sprintf("value%d", i%7),
			Condition: condition.NewTextCondition(
				condition.NewKeyCondition(
					condition.NewCondition(false), "", fmt.Sprintf("key%d", i%4),
				),
				fmt.Sprintf("pattern%d", i%5), false,
			),
		}
		_, err := svc.Create(ctx, fmt.Sprintf("acc%d", i%3), fmt.Sprintf("user%d", i%3), req)
		require.Nil(t, err)
	}
	//
	cases := map[string]struct {
		query    subscription.QueryOwn
		pageSize int
		err      error
	}{
		"key0/value0 -> 3 subs": {
			query: subscription.QueryOwn{
				Limit:   100,
				GroupId: "acc0",
				UserId:  "user0",
			},
			pageSize: 34,
		},
		"key1/value2 -> 3 subs": {
			query: subscription.QueryOwn{
				Limit:   10,
				GroupId: "acc1",
				UserId:  "user1",
			},
			pageSize: 10,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			page, err := svc.SearchOwn(ctx, c.query, "")
			if c.err == nil {
				require.Nil(t, err)
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
			case condition.TextCondition:
				equal = false
			default:
				equal = false
			}
		case condition.TextCondition:
			switch bt := b.(type) {
			case condition.GroupCondition:
				equal = false
			case condition.TextCondition:
				equal = at.GetKey() == bt.GetKey() && at.GetTerm() == bt.GetTerm()
			default:
				equal = false
			}
		}
	}
	return equal
}
