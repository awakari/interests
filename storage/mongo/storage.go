package mongo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	storageImpl struct {
		conn *mongo.Client
		db   *mongo.Database
		coll *mongo.Collection
	}
)

var (
	indices = []mongo.IndexModel{
		// id should be unique
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
					Key:   attrAcc,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
		// query by id (cursor) and kiwi
		{
			Keys: bson.D{
				{
					Key:   attrId,
					Value: 1,
				},
				{
					Key:   attrPrio,
					Value: 1,
				},
				{
					Key:   attrKiwis + "." + kiwiConditionAttrKey,
					Value: 1,
				},
				{
					Key:   attrKiwis + "." + kiwiConditionAttrPattern,
					Value: 1,
				},
				{
					Key:   attrKiwis + "." + kiwiConditionAttrPartial,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(false).
				SetSparse(true),
		},
	}
	optsSrvApi     = options.ServerAPI(options.ServerAPIVersion1)
	dataProjection = bson.D{
		{
			Key:   attrDescr,
			Value: 1,
		},
		{
			Key:   attrPrio,
			Value: 1,
		},
		{
			Key:   attrCond,
			Value: 1,
		},
	}
	optsRead = options.
			FindOne().
			SetProjection(dataProjection)
	optsDelete = options.
			FindOneAndDelete().
			SetProjection(dataProjection)
	idsProjection = bson.D{
		{
			Key:   attrId,
			Value: 1,
		},
	}
	searchByKiwiSortProjection = bson.D{
		{
			Key:   attrPrio,
			Value: -1,
		},
		{
			Key:   attrId,
			Value: -1,
		},
	}
	searchByKiwiProjection = bson.D{
		{
			Key:   attrId,
			Value: 1,
		},
		{
			Key:   attrAcc,
			Value: 1,
		},
		{
			Key:   attrPrio,
			Value: 1,
		},
		{
			Key:   attrCond,
			Value: 1,
		},
		{
			Key:   attrKiwis,
			Value: 1,
		},
	}
)

func NewStorage(ctx context.Context, cfgDb config.Db) (s storage.Storage, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
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

func (s storageImpl) Create(ctx context.Context, acc string, sd subscription.Data) (id string, err error) {
	md := sd.Metadata
	recCondition, recKiwis := encodeCondition(sd.Condition)
	rec := subscriptionWrite{
		Id:          uuid.NewString(),
		Account:     acc,
		Description: md.Description,
		Priority:    md.Priority,
		Condition:   recCondition,
		Kiwis:       recKiwis,
	}
	_, err = s.coll.InsertOne(ctx, rec)
	if err != nil {
		err = fmt.Errorf("%w: failed to insert: %s", storage.ErrInternal, err)
	} else {
		id = rec.Id
	}
	return
}

func (s storageImpl) Read(ctx context.Context, id, acc string) (sd subscription.Data, err error) {
	q := bson.M{
		attrId:  id,
		attrAcc: acc,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOne(ctx, q, optsRead)
	sd, err = decodeSingleResult(id, acc, result)
	return
}

func decodeSingleResult(id, acc string, result *mongo.SingleResult) (sd subscription.Data, err error) {
	err = result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = fmt.Errorf("%w: id=%s, acc=%s", storage.ErrNotFound, id, acc)
		} else {
			err = fmt.Errorf("%w: failed to find by id: %s, acc: %s, %s", storage.ErrInternal, id, acc, err)
		}
	} else {
		var rec subscriptionRec
		err = result.Decode(&rec)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode, id=%s, acc=%s, %s", storage.ErrInternal, id, acc, err)
		} else {
			err = rec.decodeSubscriptionData(&sd)
		}
	}
	return
}

func (s storageImpl) UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error) {
	q := bson.M{
		attrId:  id,
		attrAcc: acc,
	}
	u := bson.M{
		"$set": bson.M{
			attrDescr: md.Description,
			attrPrio:  md.Priority,
		},
	}
	var result *mongo.UpdateResult
	result, err = s.coll.UpdateOne(ctx, q, u)
	if err != nil {
		err = fmt.Errorf("%w: failed to update metadata, id: %s, err: %s", storage.ErrInternal, id, err)
	} else if result.ModifiedCount < 1 {
		err = fmt.Errorf("%w: not found, id: %s, acc: %s", storage.ErrNotFound, id, acc)
	}
	return
}

func (s storageImpl) Delete(ctx context.Context, id, acc string) (sd subscription.Data, err error) {
	q := bson.M{
		attrId:  id,
		attrAcc: acc,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOneAndDelete(ctx, q, optsDelete)
	sd, err = decodeSingleResult(id, acc, result)
	return
}

func (s storageImpl) SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error) {
	dbQuery := bson.M{
		attrId: bson.M{
			"$gt": cursor,
		},
		attrAcc: q.Account,
	}
	opts := options.
		Find().
		SetLimit(int64(q.Limit)).
		SetProjection(idsProjection).
		SetShowRecordID(false).
		SetSort(idsProjection)
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

func (s storageImpl) SearchByKiwi(ctx context.Context, q storage.KiwiQuery, cursor subscription.ConditionMatchKey) (page []subscription.ConditionMatch, err error) {
	dbQuery := bson.M{
		attrKiwis + "." + kiwiConditionAttrKey:     q.Key,
		attrKiwis + "." + kiwiConditionAttrPattern: q.Pattern,
		attrKiwis + "." + kiwiConditionAttrPartial: q.Partial,
		attrPrio: bson.M{
			"$gt": 0,
		},
	}
	if cursor.Id != "" {
		dbQuery["$or"] = []bson.M{
			{
				attrPrio: bson.M{
					"$lt": cursor.Priority,
				},
			},
			{
				attrPrio: cursor.Priority,
				attrId: bson.M{
					"$gt": cursor.Id,
				},
			},
		}
	}
	opts := options.
		Find().
		SetLimit(int64(q.Limit)).
		SetProjection(searchByKiwiProjection).
		SetShowRecordID(false).
		SetSort(searchByKiwiSortProjection)
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%v, cursor=%+v, %s", storage.ErrInternal, dbQuery, cursor, err)
	} else {
		defer cur.Close(ctx)
		var recs []subscriptionRec
		err = cur.All(ctx, &recs)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode: %s", storage.ErrInternal, err)
		} else {
			for _, rec := range recs {
				var cm subscription.ConditionMatch
				err = rec.decodeSubscriptionConditionMatch(&cm)
				var condId string
				for _, kiwi := range rec.Kiwis {
					if kiwi.Key == q.Key && kiwi.Pattern == q.Pattern && kiwi.Partial == q.Partial {
						condId = kiwi.Id
					}
				}
				cm.ConditionId = condId
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
