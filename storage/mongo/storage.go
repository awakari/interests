package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/config"
	"github.com/awakari/subscriptions/model"
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
		// name should be unique
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
		// query by name and kiwi
		{
			Keys: bson.D{
				{
					Key:   attrId,
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
	idsProjection = bson.D{
		{
			Key:   attrId,
			Value: 1,
		},
	}
	searchProjection = bson.D{
		{
			Key:   attrId,
			Value: 1,
		},
		{
			Key:   attrMetadata,
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

func (s storageImpl) Create(ctx context.Context, sd model.SubscriptionData) (id string, err error) {
	recCondition, recKiwis := encodeCondition(sd.Condition)
	rec := subscriptionWrite{
		Id:        uuid.NewString(),
		Metadata:  sd.Metadata,
		Routes:    sd.Routes,
		Condition: recCondition,
		Kiwis:     recKiwis,
	}
	_, err = s.coll.InsertOne(ctx, rec)
	if err == nil {
		id = rec.Id
	} else if mongo.IsDuplicateKeyError(err) {
		err = fmt.Errorf("%w: %s", storage.ErrConflict, err)
	}
	return
}

func (s storageImpl) Read(ctx context.Context, id string) (sd model.SubscriptionData, err error) {
	q := bson.M{
		attrId: id,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOne(ctx, q, optsRead)
	sd, err = decodeSingleResult(id, result)
	return
}

func decodeSingleResult(id string, result *mongo.SingleResult) (sd model.SubscriptionData, err error) {
	err = result.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = fmt.Errorf("%w: id=%s", storage.ErrNotFound, id)
		} else {
			err = fmt.Errorf("%w: failed to find by id: %s, %s", storage.ErrInternal, id, err)
		}
	} else {
		var rec subscription
		err = result.Decode(&rec)
		if err != nil {
			err = fmt.Errorf("%w: failed to decode, id=%s, %s", storage.ErrInternal, id, err)
		} else {
			err = rec.decodeSubscriptionData(&sd)
		}
	}
	return
}

func (s storageImpl) Delete(ctx context.Context, id string) (sd model.SubscriptionData, err error) {
	q := bson.M{
		attrId: id,
	}
	var result *mongo.SingleResult
	result = s.coll.FindOneAndDelete(ctx, q)
	sd, err = decodeSingleResult(id, result)
	return
}

func (s storageImpl) SearchByKiwi(ctx context.Context, q storage.KiwiQuery, cursor string) (page []model.Subscription, err error) {
	dbQuery := bson.M{
		attrKiwis + "." + kiwiConditionAttrKey:     q.Key,
		attrKiwis + "." + kiwiConditionAttrPattern: q.Pattern,
		attrKiwis + "." + kiwiConditionAttrPartial: q.Partial,
	}
	return s.search(ctx, dbQuery, q.Limit, cursor)
}

func (s storageImpl) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []model.Subscription, err error) {
	dbQuery := bson.M{}
	for k, v := range q.Metadata {
		dbQuery[attrMetadata+"."+k] = v
	}
	return s.search(ctx, dbQuery, q.Limit, cursor)
}

func (s storageImpl) search(ctx context.Context, dbQuery bson.M, limit uint32, cursor string) (page []model.Subscription, err error) {
	dbQuery[attrId] = bson.M{
		"$gt": cursor,
	}
	opts := options.
		Find().
		SetLimit(int64(limit)).
		SetProjection(searchProjection).
		SetShowRecordID(false).
		SetSort(idsProjection)
	var cur *mongo.Cursor
	cur, err = s.coll.Find(ctx, dbQuery, opts)
	if err != nil {
		err = fmt.Errorf("%w: failed to find: query=%v, cursor=%s, %s", storage.ErrInternal, dbQuery, cursor, err)
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
