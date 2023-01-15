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
				{attrName, 1},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
		// TODO query by kiwi
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
	recCondition, recKiwiConditions := encodeCondition(sub.Condition)
	rec := subscriptionWrite{
		Name:           sub.Name,
		Description:    sub.Description,
		Routes:         sub.Routes,
		Condition:      recCondition,
		KiwiConditions: recKiwiConditions,
	}
	_, err = s.coll.InsertOne(ctx, rec)
	if mongo.IsDuplicateKeyError(err) {
		err = fmt.Errorf("%w: %s", storage.ErrConflict, err)
	}
	return
}

func encodeCondition(src model.Condition) (dst Condition, kiwiConditions []kiwiCondition) {
	bc := ConditionBase{
		Not: src.IsNot(),
	}
	switch c := src.(type) {
	case model.GroupCondition:
		var group []Condition
		for _, childSrc := range c.GetGroup() {
			childDst, childKiwiConditions := encodeCondition(childSrc)
			group = append(group, childDst)
			kiwiConditions = append(kiwiConditions, childKiwiConditions...)
		}
		dst = groupCondition{
			Base:  bc,
			Group: group,
			Logic: c.GetLogic(),
		}
	case model.KiwiCondition:
		p := c.GetPattern()
		kc := kiwiCondition{
			Base:    bc,
			Key:     c.GetKey(),
			Partial: c.IsPartial(),
			ValuePattern: pattern{
				Code: p.Code,
				Src:  p.Src,
			},
		}
		kiwiConditions = append(kiwiConditions, kc)
		dst = kc
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

func (s storageImpl) Search(ctx context.Context, q storage.KiwiQuery, cursor string) (page []model.Subscription, err error) {
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
