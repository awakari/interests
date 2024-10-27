package mongo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type storageImpl struct {
	conn *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
}

const countUsersUnique = "countUsersUnique"

var timeZero = time.Time{}.UTC()
var (
	indices = []mongo.IndexModel{
		// external id should be unique
		{
			Keys: bson.D{
				{
					Key:   attrId,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
		// read/query by id (cursor) and account
		{
			Keys: bson.D{
				{
					Key:   attrId,
					Value: 1,
				},
				{
					Key:   attrFollowers,
					Value: 1,
				},
				{
					Key:   attrGroupId,
					Value: 1,
				},
				{
					Key:   attrUserId,
					Value: "hashed",
				},
			},
			Options: options.
				Index().
				SetSparse(true).
				SetUnique(false),
		},
		// query by enabled flag, expires and condition id
		{
			Keys: bson.D{
				{
					Key:   attrEnabled,
					Value: 1,
				},
				{
					Key:   attrExpires,
					Value: 1,
				},
				{
					Key:   attrCondIds,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(false),
		},
	}
	projId = bson.D{
		{
			Key:   attrId,
			Value: 1,
		},
	}
	projIdDesc = bson.D{
		{
			Key:   attrId,
			Value: -1,
		},
	}
	projFollowersAsc = bson.D{
		{
			Key:   attrFollowers,
			Value: 1,
		},
		{
			Key:   attrId,
			Value: 1,
		},
	}
	projFollowersDesc = bson.D{
		{
			Key:   attrFollowers,
			Value: -1,
		},
		{
			Key:   attrId,
			Value: -1,
		},
	}
	projData = bson.D{
		{
			Key:   attrDescr,
			Value: 1,
		},
		{
			Key:   attrEnabled,
			Value: 1,
		},
		{
			Key:   attrExpires,
			Value: 1,
		},
		{
			Key:   attrCond,
			Value: 1,
		},
		{
			Key:   attrCreated,
			Value: 1,
		},
		{
			Key:   attrUpdated,
			Value: 1,
		},
		{
			Key:   attrResult,
			Value: 1,
		},
		{
			Key:   attrPublic,
			Value: 1,
		},
		{
			Key:   attrFollowers,
			Value: 1,
		},
		{
			Key:   attrGroupId,
			Value: 1,
		},
		{
			Key:   attrUserId,
			Value: 1,
		},
	}
	projSearchByCondId = bson.D{
		{
			Key:   attrId,
			Value: 1,
		},
		{
			Key:   attrCond,
			Value: 1,
		},
		{
			Key:   attrCondIds,
			Value: 1,
		},
	}
	optsSrvApi = options.
			ServerAPI(options.ServerAPIVersion1)
	optsRead = options.
			FindOne().
			SetProjection(projData)
	optsUpdate = options.
			FindOneAndUpdate().
			SetProjection(projData).
			SetReturnDocument(options.Before)
	optsDelete = options.
			FindOneAndDelete().
			SetProjection(projData)
	optsSearchByCond = options.
				Find().
				SetProjection(projSearchByCondId).
				SetShowRecordID(false).
				SetSort(projId)
	pipelineCountUsersUniq = mongo.Pipeline{
		bson.D{{
			"$group",
			bson.D{{
				"_id",
				"$userId",
			}},
		}},
		bson.D{{
			"$count",
			countUsersUnique,
		}},
	}
)

func NewStorage(ctx context.Context, cfgDb config.DbConfig) (s storage.Storage, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi)
	if cfgDb.Tls.Enabled {
		clientOpts = clientOpts.SetTLSConfig(&tls.Config{InsecureSkipVerify: cfgDb.Tls.Insecure})
	}
	if len(cfgDb.UserName) > 0 {
		auth := options.Credential{
			Username:    cfgDb.UserName,
			Password:    cfgDb.Password,
			PasswordSet: len(cfgDb.Password) > 0,
		}
		clientOpts = clientOpts.SetAuth(auth)
	}
	conn, err := mongo.Connect(ctx, clientOpts)
	var stor storageImpl
	if err == nil {
		db := conn.Database(cfgDb.Name)
		coll := db.Collection(cfgDb.Table.Name)
		stor.conn = conn
		stor.db = db
		stor.coll = coll
		_, err = stor.ensureIndices(ctx)
	}
	if err == nil && cfgDb.Table.Shard {
		err = stor.shardCollection(ctx)
	}
	if err == nil {
		s = stor
	}
	if err != nil {
		err = fmt.Errorf("%w: %s", storage.ErrInternal, err)
	}
	return
}

func (s storageImpl) ensureIndices(ctx context.Context) ([]string, error) {
	return s.coll.Indexes().CreateMany(ctx, indices)
}

func (s storageImpl) shardCollection(ctx context.Context) (err error) {
	adminDb := s.conn.Database("admin")
	cmd := bson.D{
		{
			Key:   "shardCollection",
			Value: fmt.Sprintf("%s.%s", s.db.Name(), s.coll.Name()),
		},
		{
			Key: "key",
			Value: bson.D{
				{
					Key:   attrId,
					Value: "hashed",
				},
			},
		},
	}
	err = adminDb.RunCommand(ctx, cmd).Err()
	return
}

func (s storageImpl) Close() error {
	return s.conn.Disconnect(context.TODO())
}

func (s storageImpl) Create(ctx context.Context, id, groupId, userId string, sd subscription.Data) (err error) {
	recCond, condIds := encodeCondition(sd.Condition)
	rec := subscriptionWrite{
		Id:          id,
		GroupId:     groupId,
		UserId:      userId,
		Description: sd.Description,
		Enabled:     sd.Enabled,
		Expires:     sd.Expires.UTC(),
		Created:     sd.Created.UTC(),
		Updated:     sd.Updated.UTC(),
		Public:      sd.Public,
		Followers:   sd.Followers,
		Condition:   recCond,
		CondIds:     condIds,
	}
	_, err = s.coll.InsertOne(ctx, rec)
	switch {
	case mongo.IsDuplicateKeyError(err):
		err = fmt.Errorf("%w: id already in use: %s", storage.ErrConflict, id)
	case err != nil:
		err = fmt.Errorf("%w: failed to insert: %s", storage.ErrInternal, err)
	}
	return
}

func (s storageImpl) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	q := bson.M{
		attrId: id,
		"$or": []bson.M{
			{
				attrGroupId: groupId,
				attrUserId:  userId,
			},
			{
				attrPublic: true,
			},
		},
	}
	var result *mongo.SingleResult
	result = s.coll.FindOne(ctx, q, optsRead)
	sd, err = decodeSingleResult(id, groupId, userId, result)
	return
}

func decodeSingleResult(id, groupId, userId string, result *mongo.SingleResult) (sd subscription.Data, err error) {
	err = result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = fmt.Errorf("%w: id=%s, acc=%s/%s", storage.ErrNotFound, id, groupId, userId)
		} else {
			err = fmt.Errorf("%w: failed to find by id: %s, acc: %s/%s, %s", storage.ErrInternal, id, groupId, userId, err)
		}
	} else {
		var rec subscriptionRec
		err = result.Decode(&rec)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode, id=%s, acc=%s/%s, %s", storage.ErrInternal, id, groupId, userId, err)
		} else {
			err = rec.decodeSubscriptionData(&sd)
			if groupId == rec.GroupId && userId == rec.UserId {
				sd.Own = true
			}
		}
	}
	return
}

func (s storageImpl) Update(ctx context.Context, id, groupId, userId string, d subscription.Data) (prev subscription.Data, err error) {
	q := bson.M{
		attrId:      id,
		attrGroupId: groupId,
		attrUserId:  userId,
	}
	cond, condIds := encodeCondition(d.Condition)
	u := bson.M{
		"$set": bson.M{
			attrDescr:   d.Description,
			attrEnabled: d.Enabled,
			attrExpires: d.Expires.UTC(),
			attrUpdated: d.Updated.UTC(),
			attrPublic:  d.Public,
			attrCond:    cond,
			attrCondIds: condIds,
		},
	}
	var result *mongo.SingleResult
	result = s.coll.FindOneAndUpdate(ctx, q, u, optsUpdate)
	err = result.Err()
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		err = fmt.Errorf("%w: not found, id: %s, acc: %s/%s", storage.ErrNotFound, id, groupId, userId)
	case err != nil:
		err = fmt.Errorf("%w: failed to update subscription, id: %s, err: %s", storage.ErrInternal, id, err)
	default:
		prev, err = decodeSingleResult(id, groupId, userId, result)
	}
	return
}

func (s storageImpl) UpdateFollowers(ctx context.Context, id string, count int64) (err error) {
	q := bson.M{
		attrId: id,
	}
	u := bson.M{
		"$set": bson.M{
			attrFollowers: count,
		},
	}
	var result *mongo.UpdateResult
	result, err = s.coll.UpdateOne(ctx, q, u)
	switch {
	case err == nil && result.MatchedCount < 1:
		err = fmt.Errorf("%w: not found, id: %s", storage.ErrNotFound, id)
	case err != nil:
		err = fmt.Errorf("%w: failed to update subscription, id: %s, err: %s", storage.ErrInternal, id, err)
	}
	return
}

func (s storageImpl) UpdateResultTime(ctx context.Context, id string, last time.Time) (err error) {
	q := bson.M{
		attrId: id,
	}
	u := bson.M{
		"$set": bson.M{
			attrResult: last.UTC(),
		},
	}
	var result *mongo.UpdateResult
	result, err = s.coll.UpdateOne(ctx, q, u)
	switch {
	case err == nil && result.MatchedCount < 1:
		err = fmt.Errorf("%w: not found, id: %s", storage.ErrNotFound, id)
	case err != nil:
		err = fmt.Errorf("%w: failed to update subscription, id: %s, err: %s", storage.ErrInternal, id, err)
	}
	return
}

func (s storageImpl) SetEnabledBatch(ctx context.Context, ids []string, enabled bool) (n int64, err error) {
	q := bson.M{
		attrId: bson.M{
			"$in": ids,
		},
	}
	u := bson.M{
		"$set": bson.M{
			attrEnabled: enabled,
		},
	}
	var result *mongo.UpdateResult
	result, err = s.coll.UpdateMany(ctx, q, u)
	switch {
	case err == nil:
		n = result.ModifiedCount
	default:
		err = fmt.Errorf("%w: failed to update subscription, ids: %s, err: %s", storage.ErrInternal, ids, err)
	}
	return
}

func (s storageImpl) Delete(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	q := bson.M{
		attrId:      id,
		attrGroupId: groupId,
		attrUserId:  userId,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOneAndDelete(ctx, q, optsDelete)
	sd, err = decodeSingleResult(id, groupId, userId, result)
	return
}

func (s storageImpl) Search(ctx context.Context, q subscription.Query, cursor subscription.Cursor) (ids []string, err error) {
	opts := options.
		Find().
		SetLimit(int64(q.Limit)).
		SetProjection(projId).
		SetShowRecordID(false)
	dbQuery := bson.M{}
	switch q.Public {
	case true:
		dbQuery["$or"] = []bson.M{
			{
				attrGroupId: q.GroupId,
				attrUserId:  q.UserId,
			},
			{
				attrPublic: true,
			},
		}
	default:
		dbQuery[attrGroupId] = q.GroupId
		dbQuery[attrUserId] = q.UserId
	}
	switch q.Sort {
	case subscription.SortFollowers:
		switch q.Order {
		case subscription.OrderDesc:
			switch cursor.Followers {
			case 0:
				dbQuery = bson.M{
					"$and": []bson.M{
						dbQuery,
						{
							"$or": []bson.M{
								{
									attrFollowers: cursor.Followers,
								},
								{
									attrFollowers: bson.M{
										"$exists": false,
									},
								},
							},
						},
						{
							attrId: bson.M{
								"$lt": cursor.Id,
							},
						},
					},
				}
			default:
				dbQuery = bson.M{
					"$and": []bson.M{
						dbQuery,
						{
							"$or": []bson.M{
								{
									"$or": []bson.M{
										{
											attrFollowers: bson.M{
												"$lt": cursor.Followers,
											},
										},
										{
											attrFollowers: bson.M{
												"$exists": false,
											},
										},
									},
								},
								{
									"$and": []bson.M{
										{
											attrFollowers: cursor.Followers,
										},
										{
											attrId: bson.M{
												"$lt": cursor.Id,
											},
										},
									},
								},
							},
						},
					},
				}
			}
			opts = opts.SetSort(projFollowersDesc)
		default:
			switch cursor.Followers {
			case 0:
				dbQuery = bson.M{
					"$and": []bson.M{
						dbQuery,
						{
							"$or": []bson.M{
								{
									"$or": []bson.M{
										{
											attrFollowers: bson.M{
												"$gt": cursor.Followers,
											},
										},
										{
											attrFollowers: bson.M{
												"$exists": true,
											},
										},
									},
								},
								{
									"$and": []bson.M{
										{
											"$or": []bson.M{
												{
													attrFollowers: cursor.Followers,
												},
												{
													attrFollowers: bson.M{
														"$exists": false,
													},
												},
											},
										},
										{
											attrId: bson.M{
												"$gt": cursor.Id,
											},
										},
									},
								},
							},
						},
					},
				}
			default:
				dbQuery = bson.M{
					"$and": []bson.M{
						dbQuery,
						{
							"$or": []bson.M{
								{
									attrFollowers: bson.M{
										"$gt": cursor.Followers,
									},
								},
								{
									"$and": []bson.M{
										{
											attrFollowers: cursor.Followers,
										},
										{
											attrId: bson.M{
												"$gt": cursor.Id,
											},
										},
									},
								},
							},
						},
					},
				}
			}
			opts = opts.SetSort(projFollowersAsc)
		}
	default:
		switch q.Order {
		case subscription.OrderDesc:
			dbQuery[attrId] = bson.M{
				"$lt": cursor.Id,
			}
			opts = opts.SetSort(projIdDesc)
		default:
			dbQuery[attrId] = bson.M{
				"$gt": cursor.Id,
			}
			opts = opts.SetSort(projId)
		}
	}
	dbQuery[attrDescr] = bson.M{
		"$regex": q.Pattern,
	}
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%v, cursor=%v, %s", storage.ErrInternal, dbQuery, cursor, err)
	} else {
		defer cur.Close(ctx)
		var recs []subscriptionRec
		err = cur.All(ctx, &recs)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode: %s", storage.ErrInternal, err)
		} else {
			for _, rec := range recs {
				ids = append(ids, rec.Id)
			}
		}
	}
	return
}

func (s storageImpl) SearchByCondition(ctx context.Context, q subscription.QueryByCondition, cursor string) (page []subscription.ConditionMatch, err error) {
	dbQuery := bson.M{
		attrId: bson.M{
			"$gt": cursor,
		},
		attrCondIds: q.CondId,
		attrEnabled: true,
		"$or": []bson.M{
			{
				attrExpires: bson.M{
					"$gt": time.Now().UTC(),
				},
			},
			{
				attrExpires: timeZero,
			},
			{
				attrExpires: bson.M{
					"$exists": false,
				},
			},
		},
	}
	opts := optsSearchByCond.SetLimit(int64(q.Limit))
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%+v, %s", storage.ErrInternal, dbQuery, err)
	} else {
		defer cur.Close(ctx)
		var recs []subscriptionRec
		err = cur.All(ctx, &recs)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode subscription record @ cursor %v: %s", storage.ErrInternal, cur.Current, err)
		} else {
			for _, rec := range recs {
				var cm subscription.ConditionMatch
				err = rec.decodeSubscriptionConditionMatch(&cm)
				if err != nil {
					err = fmt.Errorf("%w: failed to decode subscription record %v: %s", storage.ErrInternal, rec, err)
					break
				}
				page = append(page, cm)
			}
		}
	}
	return
}

func (s storageImpl) Count(ctx context.Context) (count int64, err error) {
	return s.coll.EstimatedDocumentCount(ctx)
}

func (s storageImpl) CountUsersUnique(ctx context.Context) (count int64, err error) {
	var cursor *mongo.Cursor
	cursor, err = s.coll.Aggregate(ctx, pipelineCountUsersUniq)
	var result bson.M
	if err == nil && cursor.Next(ctx) {
		err = cursor.Decode(&result)
	}
	if err == nil {
		rawCount := result[countUsersUnique]
		switch rawCount.(type) {
		case int32:
			count = int64(rawCount.(int32))
		case int64:
			count = rawCount.(int64)
		default:
			err = fmt.Errorf("%w: failed to convert result to int: %+v", storage.ErrInternal, rawCount)
		}
	}
	return
}
