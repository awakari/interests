package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/meandros-messaging/subscriptions/config"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/storage"
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
				{attrName, 1},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
		// query by matcher in excludes group
		{
			Keys: bson.D{
				{attrName, 1},
				{attrExcludes + "." + attrMatchers + "." + attrPartial, 1},
				{attrExcludes + "." + attrMatchers + "." + attrKey, 1},
				{attrExcludes + "." + attrMatchers + "." + attrPattern + "." + attrCode, 1},
			},
			Options: options.
				Index().
				SetUnique(true).
				SetSparse(true),
		},
		// query by matcher in includes group
		{
			Keys: bson.D{
				{attrName, 1},
				{attrIncludes + "." + attrMatchers + "." + attrPartial, 1},
				{attrIncludes + "." + attrMatchers + "." + attrKey, 1},
				{attrIncludes + "." + attrMatchers + "." + attrPattern + "." + attrCode, 1},
			},
			Options: options.
				Index().
				SetUnique(true).
				SetSparse(true),
		},
	}
	optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
	optsRead   = options.
			FindOne().
			SetShowRecordID(false)
	listNamesProjection = bson.D{
		{
			attrName,
			1,
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
	rec := subscription{
		Name:        sub.Name,
		Description: sub.Description,
		Routes:      sub.Routes,
		Includes:    toMatcherGroupRec(sub.Includes),
		Excludes:    toMatcherGroupRec(sub.Excludes),
	}
	_, err = s.coll.InsertOne(ctx, rec)
	if mongo.IsDuplicateKeyError(err) {
		err = fmt.Errorf("%w: %s", storage.ErrConflict, err)
	}
	return
}

func toMatcherGroupRec(mg model.MatcherGroup) (rec matcherGroup) {
	rec.All = mg.All
	var matcherRecs []matcher
	for _, m := range mg.Matchers {
		matcherRec := toMatcherRec(m)
		matcherRecs = append(matcherRecs, matcherRec)
	}
	rec.Matchers = matcherRecs
	return
}

func toMatcherRec(m model.Matcher) (rec matcher) {
	return matcher{
		Partial: m.Partial,
		Key:     m.Key,
		Pattern: pattern{
			Code: m.Pattern.Code,
			Src:  m.Pattern.Src,
		},
	}
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
			decodeSubscription(rec, &sub)
		}
	}
	return
}

func decodeSubscription(rec subscription, sub *model.Subscription) {
	sub.Name = rec.Name
	sub.Description = rec.Description
	sub.Routes = rec.Routes
	sub.Excludes.All = rec.Excludes.All
	sub.Excludes.Matchers = decodeMatchers(rec.Excludes.Matchers)
	sub.Includes.All = rec.Includes.All
	sub.Includes.Matchers = decodeMatchers(rec.Includes.Matchers)
}

func decodeMatchers(matcherRecs []matcher) (matchers []model.Matcher) {
	for _, matcherRec := range matcherRecs {
		m := model.Matcher{
			MatcherData: model.MatcherData{
				Key: matcherRec.Key,
				Pattern: model.Pattern{
					Code: matcherRec.Pattern.Code,
					Src:  matcherRec.Pattern.Src,
				},
			},
			Partial: matcherRec.Partial,
		}
		matchers = append(matchers, m)
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
		SetProjection(listNamesProjection).
		SetShowRecordID(false).
		SetSort(listNamesProjection)
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

func (s storageImpl) Search(ctx context.Context, q storage.Query, cursor string) (page []model.Subscription, err error) {
	dbQuery := bson.M{
		attrName: bson.M{
			"$gt": cursor,
		},
	}
	var attrMatcherGroup string
	if q.InExcludes {
		attrMatcherGroup = attrExcludes
	} else {
		attrMatcherGroup = attrIncludes
	}
	dbQuery[attrMatcherGroup+"."+attrMatchers+"."+attrPartial] = q.Matcher.Partial
	dbQuery[attrMatcherGroup+"."+attrMatchers+"."+attrKey] = q.Matcher.Key
	dbQuery[attrMatcherGroup+"."+attrMatchers+"."+attrPattern+"."+attrCode] = q.Matcher.Pattern.Code
	opts := options.
		Find().
		SetLimit(int64(q.Limit)).
		SetShowRecordID(false).
		SetSort(listNamesProjection)
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
				decodeSubscription(rec, &sub)
				page = append(page, sub)
			}
		}
	}
	return
}
