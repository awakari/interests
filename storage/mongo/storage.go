package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
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
		// name should be unique
		{
			Keys: bson.D{
				{
					Key:   attrName,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
		// query by name and kiwi
		{
			Keys: bson.D{
				{
					Key:   attrName,
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
	optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
	optsRead   = options.
			FindOne().
			SetShowRecordID(false)
	namesProjection = bson.D{
		{
			Key:   attrName,
			Value: 1,
		},
	}
	searchProjection = bson.D{
		{
			Key:   attrName,
			Value: 1,
		},
		{
			Key:   attrRoutes,
			Value: 1,
		},
		{
			Key:   attrCondition,
			Value: 1,
		},
	}
)

func NewStorage(ctx context.Context, cfgDb config.Db) (s storage.Storage, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi)
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

func (s storageImpl) Create(ctx context.Context, sub model.Subscription) (err error) {
	recCondition, recKiwis := encodeCondition(sub.Condition)
	rec := subscriptionWrite{
		Name:        sub.Name,
		Description: sub.Description,
		Routes:      sub.Routes,
		Condition:   recCondition,
		Kiwis:       recKiwis,
	}
	_, err = s.coll.InsertOne(ctx, rec)
	if mongo.IsDuplicateKeyError(err) {
		err = fmt.Errorf("%w: %s", storage.ErrConflict, err)
	}
	return
}

func (s storageImpl) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	q := bson.M{
		attrName: name,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOne(ctx, q, optsRead)
	sub, err = decodeSingleResult(name, result)
	return
}

func decodeSingleResult(name string, result *mongo.SingleResult) (sub model.Subscription, err error) {
	err = result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = fmt.Errorf("%w: name=%s", storage.ErrNotFound, name)
		} else {
			err = fmt.Errorf("%w: failed to find by name: %s, %s", storage.ErrInternal, name, err)
		}
	} else {
		var rec subscription
		err = result.Decode(&rec)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode, name=%s, %s", storage.ErrInternal, name, err)
		} else {
			err = rec.decodeSubscription(&sub)
		}
	}
	return
}

func (s storageImpl) Delete(ctx context.Context, name string) (sub model.Subscription, err error) {
	q := bson.M{
		attrName: name,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOneAndDelete(ctx, q)
	sub, err = decodeSingleResult(name, result)
	return
}

func (s storageImpl) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	q := bson.M{
		attrName: bson.M{
			"$gt": cursor,
		},
	}
	opts := options.
		Find().
		SetLimit(int64(limit)).
		SetProjection(namesProjection).
		SetShowRecordID(false).
		SetSort(namesProjection)
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, q, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to list: limit=%d, cursor=%s, %s", storage.ErrInternal, limit, cursor, err)
	} else {
		defer cur.Close(ctx)
		var recs []subscription
		err = cur.All(ctx, &recs)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode: %s", storage.ErrInternal, err)
		} else {
			for _, rec := range recs {
				page = append(page, rec.Name)
			}
		}
	}
	return
}

func (s storageImpl) SearchByKiwi(ctx context.Context, q storage.KiwiQuery, cursor string) (page []model.Subscription, err error) {
	dbQuery := bson.M{
		attrName: bson.M{
			"$gt": cursor,
		},
		attrKiwis + "." + kiwiConditionAttrKey:     q.Key,
		attrKiwis + "." + kiwiConditionAttrPattern: q.Pattern,
		attrKiwis + "." + kiwiConditionAttrPartial: q.Partial,
	}
	opts := options.
		Find().
		SetLimit(int64(q.Limit)).
		SetProjection(searchProjection).
		SetShowRecordID(false).
		SetSort(namesProjection)
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%v, cursor=%s, %s", storage.ErrInternal, q, cursor, err)
	} else {
		defer cur.Close(ctx)
		var recs []subscription
		err = cur.All(ctx, &recs)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode: %s", storage.ErrInternal, err)
		} else {
			for _, rec := range recs {
				var sub model.Subscription
				err = rec.decodeSubscription(&sub)
				if err != nil {
					err = fmt.Errorf("%w: failed to decode subscription record %v: %s", storage.ErrInternal, rec, err)
					break
				}
				page = append(page, sub)
			}
		}
	}
	return
}
