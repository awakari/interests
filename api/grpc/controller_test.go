package grpc

import (
	"context"
	"fmt"
	"github.com/awakari/interests/api/grpc/common"
	"github.com/awakari/interests/model/interest"
	"github.com/awakari/interests/storage"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"os"
	"testing"
	"time"
)

const port = 50051

var log = slog.Default()

func TestMain(m *testing.M) {
	stor := storage.NewStorageMock(make(map[string]interest.Data))
	stor = storage.NewLoggingMiddleware(stor, log)
	go func() {
		err := Serve(stor, port)
		if err != nil {
			log.Error(err.Error())
		}
	}()
	code := m.Run()
	os.Exit(code)
}

func TestServiceController_Create(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		md      []string
		expires *timestamppb.Timestamp
		cond    *Condition
		public  bool
		id      string
		err     error
	}{
		"empty id": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			expires: timestamppb.New(time.Now()),
			public:  true,
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.InvalidArgument, "empty interest id"),
		},
		"ok2": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Not: false,
				Cond: &Condition_Gc{
					Gc: &GroupCondition{
						Logic: common.GroupLogic_And,
						Group: []*Condition{
							{
								Not: true,
								Cond: &Condition_Tc{
									Tc: &TextCondition{
										Key:  "key0",
										Term: "pattern0",
									},
								},
							},
							{
								Cond: &Condition_Sc{
									Sc: &SemanticCondition{
										Query: "lorem ipsum...",
									},
								},
							},
							{
								Cond: &Condition_Nc{
									Nc: &NumberCondition{
										Key: "key3",
										Op:  Operation_Gt,
										Val: 1.23,
									},
								},
							},
						},
					},
				},
			},
			id: "interest2",
		},
		"fail": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			id:  "fail",
			err: status.Error(codes.Internal, "internal interest storage failure"),
		},
		"conflict": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			id:  "conflict",
			err: status.Error(codes.AlreadyExists, "interest id is already in use"),
		},
		"empty group": {
			md: []string{
				"x-awakari-group-id", "",
				"X-Awakari-User-ID", "user0",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
		"empty user": {
			md: []string{
				"x-awakari-group-id", "group0",
				"X-Awakari-User-ID", "",
			},
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-user-id in request metadata"),
		},
		"no auth info": {
			cond: &Condition{
				Cond: &Condition_Tc{
					Tc: &TextCondition{},
				},
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := metadata.AppendToOutgoingContext(context.TODO(), c.md...)
			_, err = client.Create(ctx, &CreateRequest{
				Id:          c.id,
				Description: k,
				Expires:     c.expires,
				Cond:        c.cond,
				Public:      c.public,
			})
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestServiceController_Read(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth     bool
		internal bool
		sub      *ReadResponse
		err      error
	}{
		"ok": {
			auth: true,
			sub: &ReadResponse{
				Description:  "description",
				EnabledSince: timestamppb.New(time.Date(2025, 2, 1, 7, 20, 45, 0, time.UTC)),
				Expires:      timestamppb.New(time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC)),
				Created:      timestamppb.New(time.Date(2024, 4, 9, 7, 3, 25, 0, time.UTC)),
				Updated:      timestamppb.New(time.Date(2024, 4, 9, 7, 3, 35, 0, time.UTC)),
				Result:       timestamppb.New(time.Date(2024, 4, 9, 7, 3, 45, 0, time.UTC)),
				Enabled:      true,
				Public:       true,
				Followers:    42,
				Cond: &Condition{
					Not: false,
					Cond: &Condition_Gc{
						Gc: &GroupCondition{
							Logic: common.GroupLogic_And,
							Group: []*Condition{
								{
									Not: false,
									Cond: &Condition_Tc{
										Tc: &TextCondition{
											Key:  "key0",
											Term: "pattern0",
										},
									},
								},
								{
									Not: true,
									Cond: &Condition_Sc{
										Sc: &SemanticCondition{
											Query: "lorem ipsum...",
										},
									},
								},
								{
									Cond: &Condition_Nc{
										Nc: &NumberCondition{
											Key: "key2",
											Op:  Operation_Eq,
											Val: 42,
										},
									},
								},
							},
						},
					},
				},
				GroupId: "group0",
				UserId:  "user0",
			},
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal interest storage failure"),
		},
		"missing": {
			auth: true,
			err:  status.Error(codes.NotFound, "interest was not found"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
		"internal": {
			auth:     false,
			internal: true,
			sub: &ReadResponse{
				Description:  "description",
				Expires:      timestamppb.New(time.Date(2023, 10, 4, 10, 20, 45, 0, time.UTC)),
				Created:      timestamppb.New(time.Date(2024, 4, 9, 7, 3, 25, 0, time.UTC)),
				Updated:      timestamppb.New(time.Date(2024, 4, 9, 7, 3, 35, 0, time.UTC)),
				Result:       timestamppb.New(time.Date(2024, 4, 9, 7, 3, 45, 0, time.UTC)),
				Enabled:      true,
				EnabledSince: timestamppb.New(time.Date(2025, 2, 1, 7, 20, 45, 0, time.UTC)),
				Public:       true,
				Followers:    42,
				Cond: &Condition{
					Not: false,
					Cond: &Condition_Gc{
						Gc: &GroupCondition{
							Logic: common.GroupLogic_And,
							Group: []*Condition{
								{
									Not: false,
									Cond: &Condition_Tc{
										Tc: &TextCondition{
											Key:  "key0",
											Term: "pattern0",
										},
									},
								},
								{
									Not: true,
									Cond: &Condition_Sc{
										Sc: &SemanticCondition{
											Query: "lorem ipsum...",
										},
									},
								},
								{
									Cond: &Condition_Nc{
										Nc: &NumberCondition{
											Key: "key2",
											Op:  Operation_Eq,
											Val: 42,
										},
									},
								},
							},
						},
					},
				},
				GroupId: "group0",
				UserId:  "user0",
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			}
			sub, err := client.Read(ctx, &ReadRequest{
				Id:       k,
				Internal: c.internal,
			})
			if c.err == nil {
				require.Nil(t, err)
				assert.Equal(t, c.sub.Description, sub.Description)
				assert.Equal(t, c.sub.Enabled, sub.Enabled)
				assert.Equal(t, c.sub.EnabledSince, sub.EnabledSince)
				assert.Equal(t, c.sub.Expires, sub.Expires)
				assert.Equal(t, c.sub.Created, sub.Created)
				assert.Equal(t, c.sub.Updated, sub.Updated)
				assert.Equal(t, c.sub.Result, sub.Result)
				assert.Equal(t, c.sub.Public, sub.Public)
				assert.Equal(t, c.sub.Followers, sub.Followers)
				assert.Equal(t, c.sub.Cond.Not, sub.Cond.Not)
				assert.Equal(t, c.sub.Cond.GetGc().Logic, sub.Cond.GetGc().Logic)
				assert.Equal(t, len(c.sub.Cond.GetGc().GetGroup()), len(sub.Cond.GetGc().GetGroup()))
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].Not, sub.Cond.GetGc().GetGroup()[0].Not)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetTc().Key, sub.Cond.GetGc().GetGroup()[0].GetTc().Key)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[0].GetTc().Term, sub.Cond.GetGc().GetGroup()[0].GetTc().Term)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].Not, sub.Cond.GetGc().GetGroup()[1].Not)
				assert.Equal(t, c.sub.Cond.GetGc().GetGroup()[1].GetSc().Query, sub.Cond.GetGc().GetGroup()[1].GetSc().Query)
				assert.Equal(t, c.sub.Own, sub.Own)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Update(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth    bool
		descr   string
		enabled bool
		err     error
	}{
		"ok1": {
			auth: true,
		},
		"ok2": {
			auth:    true,
			descr:   "new description",
			enabled: true,
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal interest storage failure"),
		},
		"missing": {
			auth: true,
			err:  status.Error(codes.NotFound, "interest was not found"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			}
			resp, err := client.Update(ctx, &UpdateRequest{
				Id:          k,
				Description: c.descr,
				Enabled:     c.enabled,
				Expires:     timestamppb.Now(),
				Cond: &Condition{
					Not: false,
					Cond: &Condition_Gc{
						Gc: &GroupCondition{
							Logic: common.GroupLogic_And,
							Group: []*Condition{
								{
									Not: true,
									Cond: &Condition_Tc{
										Tc: &TextCondition{
											Key:  "key0",
											Term: "pattern0",
										},
									},
								},
								{
									Cond: &Condition_Tc{
										Tc: &TextCondition{
											Key:  "key1",
											Term: "pattern1",
										},
									},
								},
								{
									Cond: &Condition_Nc{
										Nc: &NumberCondition{
											Key: "key3",
											Op:  Operation_Gt,
											Val: 1.23,
										},
									},
								},
							},
						},
					},
				},
			})
			if c.err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, resp.Cond.GetGc())
				assert.Equal(t, common.GroupLogic_And, resp.Cond.GetGc().Logic)
				assert.Equal(t, "sem_0", resp.Cond.GetGc().GetGroup()[0].GetSc().Id)
				assert.Equal(t, "txt_1", resp.Cond.GetGc().GetGroup()[1].GetTc().Id)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_UpdateFollowers(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		id  string
		err error
	}{
		"ok": {},
		"fail": {
			id:  "fail",
			err: status.Error(codes.Internal, "internal interest storage failure"),
		},
		"missing": {
			id:  "missing",
			err: status.Error(codes.NotFound, "interest was not found"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			_, err = client.UpdateFollowers(ctx, &UpdateFollowersRequest{
				Id: c.id,
			})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_UpdateResultTime(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		id  string
		t   *timestamppb.Timestamp
		err error
	}{
		"ok": {
			t: &timestamppb.Timestamp{},
		},
		"invalid": {
			err: status.Error(codes.InvalidArgument, "interest  update result time missing argument"),
		},
		"fail": {
			id:  "fail",
			t:   &timestamppb.Timestamp{},
			err: status.Error(codes.Internal, "internal interest storage failure"),
		},
		"missing": {
			id:  "missing",
			t:   &timestamppb.Timestamp{},
			err: status.Error(codes.NotFound, "interest was not found"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			_, err = client.UpdateResultTime(ctx, &UpdateResultTimeRequest{
				Id:   c.id,
				Read: c.t,
			})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Delete(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth bool
		err  error
	}{
		"ok": {
			auth: true,
		},
		"fail": {
			auth: true,
			err:  status.Error(codes.Internal, "internal interest storage failure"),
		},
		"missing": {
			auth: true,
			err:  status.Error(codes.NotFound, "interest was not found"),
		},
		"no auth": {
			auth: false,
			err:  status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			}
			_, err := client.Delete(ctx, &DeleteRequest{Id: k})
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_SearchOwn(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth   bool
		cursor string
		order  Order
		err    error
		ids    []string
	}{
		"asc": {
			auth: true,
			ids: []string{
				"sub0",
				"sub1",
			},
		},
		"desc": {
			auth:  true,
			order: Order_DESC,
			ids: []string{
				"sub1",
				"sub0",
			},
		},
		"fail": {
			auth:   true,
			cursor: "fail",
			err:    status.Error(codes.Internal, "internal interest storage failure"),
		},
		"no auth": {
			auth:   false,
			cursor: "no auth",
			err:    status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			}
			resp, err := client.SearchOwn(ctx, &SearchOwnRequest{Cursor: c.cursor, Limit: 0, Order: c.order})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.ids, resp.Ids)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_Search(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		auth   bool
		cursor Cursor
		sort   Sort
		order  Order
		err    error
		ids    []string
	}{
		"asc": {
			auth: true,
			ids: []string{
				"sub0",
				"sub1",
			},
		},
		"desc": {
			auth:  true,
			order: Order_DESC,
			ids: []string{
				"sub1",
				"sub0",
			},
		},
		"desc by followers": {
			auth:  true,
			sort:  Sort_FOLLOWERS,
			order: Order_DESC,
			ids: []string{
				"sub0",
				"sub1",
			},
		},
		"fail": {
			auth: true,
			cursor: Cursor{
				Id: "fail",
			},
			err: status.Error(codes.Internal, "internal interest storage failure"),
		},
		"no auth": {
			auth: false,
			cursor: Cursor{
				Id: "no auth",
			},
			err: status.Error(codes.Unauthenticated, "missing value for x-awakari-group-id in request metadata"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			if c.auth {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			}
			resp, err := client.Search(ctx, &SearchRequest{Cursor: &c.cursor, Limit: 0, Sort: c.sort, Order: c.order})
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.ids, resp.Ids)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}

func TestServiceController_SearchByCondition(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	//
	req := &SearchByConditionRequest{
		CondId: "cond0",
		Cursor: "sub0",
		Limit:  100,
	}
	//
	cases := map[string]struct {
		count   int
		expires *timestamp.Timestamp
		err     error
	}{
		"10 seconds is more than enough to read 10_000 items": {
			count:   100,
			expires: timestamppb.New(time.Date(2025, 3, 1, 13, 4, 55, 0, time.UTC)),
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.Nil(t, err)
			defer conn.Close()
			client := NewServiceClient(conn)
			resp, err := client.SearchByCondition(context.TODO(), req)
			require.Nil(t, err)
			assert.ErrorIs(t, err, c.err)
			count := len(resp.Page)
			assert.Equal(t, c.count, count)
			assert.Equal(t, c.expires, resp.Expires)
		})
	}
}

func TestServiceController_SetEnabledBatch(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		ids          []string
		enabled      bool
		enabledSince *timestamppb.Timestamp
		n            int64
		err          error
	}{
		"ok": {
			ids: []string{
				"interest0",
				"interest1",
			},
			enabled:      true,
			enabledSince: timestamppb.New(time.Now()),
			n:            2,
		},
		"fail": {
			ids: []string{
				"fail",
				"interest1",
			},
			err: status.Error(codes.Internal, "internal interest storage failure"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.TODO()
			ctx = metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", "group0", "x-awakari-user-id", "user0")
			var resp *SetEnabledBatchResponse
			resp, err = client.SetEnabledBatch(ctx, &SetEnabledBatchRequest{
				Ids:     c.ids,
				Enabled: c.enabled,
			})
			if c.err == nil {
				assert.Equal(t, c.n, resp.N)
			}
			assert.ErrorIs(t, err, c.err)
		})
	}
}
