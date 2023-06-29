package mongo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"github.com/awakari/subscriptions/util"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type storageImpl struct {
	conn *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
}

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
					Key:   attrGroupId,
					Value: 1,
				},
				{
					Key:   attrUserId,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
		// query by enabled flag and condition id
		{
			Keys: bson.D{
				{
					Key:   attrEnabled,
					Value: 1,
				},
				{
					Key:   attrCondIds,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(false).
				SetSparse(true),
		},
	}
	optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
	projData   = bson.D{
		{
			Key:   attrDescr,
			Value: 1,
		},
		{
			Key:   attrCond,
			Value: 1,
		},
	}
	optsRead = options.
			FindOne().
			SetProjection(projData)
	optsDelete = options.
			FindOneAndDelete().
			SetProjection(projData)
	projId = bson.D{
		{
			Key:   attrId,
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
	if err != nil {
		err = fmt.Errorf("%w: %s", storage.ErrInternal, err)
	} else {
		db := conn.Database(cfgDb.Name)
		coll := db.Collection(cfgDb.Table.Name)
		stor := storageImpl{
			conn: conn,
			db:   db,
			coll: coll,
		}
		_, err = stor.ensureIndices(ctx)
		if err != nil {
			err = fmt.Errorf("%w: %s", storage.ErrInternal, err)
		} else {
			s = stor
		}
	}
	return
}

func (s storageImpl) ensureIndices(ctx context.Context) ([]string, error) {
	return s.coll.Indexes().CreateMany(ctx, indices)
}

func (s storageImpl) Close() error {
	return s.conn.Disconnect(context.TODO())
}

func (s storageImpl) Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error) {
	recCond, condIds := encodeCondition(sd.Condition)
	rec := subscriptionWrite{
		Id:          uuid.NewString(),
		GroupId:     groupId,
		UserId:      userId,
		Description: sd.Description,
		Enabled:     sd.Enabled,
		Condition:   recCond,
		CondIds:     condIds,
	}
	_, err = s.coll.InsertOne(ctx, rec)
	if err != nil {
		err = fmt.Errorf("%w: failed to insert: %s", storage.ErrInternal, err)
	} else {
		id = rec.Id
	}
	return
}

func (s storageImpl) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	q := bson.M{
		attrId:      id,
		attrGroupId: groupId,
		attrUserId:  userId,
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
		}
	}
	return
}

func (s storageImpl) Update(ctx context.Context, id, groupId, userId string, d subscription.Data) (err error) {
	q := bson.M{
		attrId:      id,
		attrGroupId: groupId,
		attrUserId:  userId,
	}
	u := bson.M{
		"$set": bson.M{
			attrDescr:   d.Description,
			attrEnabled: d.Enabled,
		},
	}
	var result *mongo.UpdateResult
	result, err = s.coll.UpdateOne(ctx, q, u)
	if err != nil {
		err = fmt.Errorf("%w: failed to update metadata, id: %s, err: %s", storage.ErrInternal, id, err)
	} else if result.ModifiedCount < 1 {
		err = fmt.Errorf("%w: not found, id: %s, acc: %s/%s", storage.ErrNotFound, id, groupId, userId)
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

func (s storageImpl) SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error) {
	dbQuery := bson.M{
		attrId: bson.M{
			"$gt": cursor,
		},
		attrGroupId: q.GroupId,
		attrUserId:  q.UserId,
	}
	opts := options.
		Find().
		SetLimit(int64(q.Limit)).
		SetProjection(projId).
		SetShowRecordID(false).
		SetSort(projId)
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%v, cursor=%s, %s", storage.ErrInternal, dbQuery, cursor, err)
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

func (s storageImpl) SearchByCondition(ctx context.Context, condId string, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	dbQuery := bson.M{
		attrCondIds: condId,
		attrEnabled: true,
	}
	opts := options.
		Find().
		SetProjection(projSearchByCondId).
		SetShowRecordID(false)
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%+v, %s", storage.ErrInternal, dbQuery, err)
	} else {
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var rec subscriptionRec
			err = cur.Decode(&rec)
			if err != nil {
				err = fmt.Errorf("%w: failed to decode subscription record @ cursor %v: %s", storage.ErrInternal, cur.Current, err)
				break
			}
			var cm subscription.ConditionMatch
			err = rec.decodeSubscriptionConditionMatch(&cm)
			if err != nil {
				err = fmt.Errorf("%w: failed to decode subscription record %v: %s", storage.ErrInternal, rec, err)
				break
			}
			err = consumeFunc(&cm)
			if err != nil {
				err = fmt.Errorf("%w: failed to consume condition match %v: %s", storage.ErrInternal, cm, err)
				break
			}
		}
	}
	return
}
