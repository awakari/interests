package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"subscriptions/storage"
)

type (
	mongoStorage struct {
		conn *mongo.Client
		db   *mongo.Database
		coll *mongo.Collection
	}
)

var (
	indices = []mongo.IndexModel{
		{
			Keys: bson.D{
				{attrExtId, 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
)

func NewStorage(ctx context.Context, uri, dbName, collName string) (storage.Storage, error) {
	clientOpts := options.
		Client().
		ApplyURI(uri).
		SetServerAPIOptions(optsSrvApi)
	conn, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	db := conn.Database(dbName)
	coll := db.Collection(collName)
	stor := mongoStorage{
		conn: conn,
		db:   db,
		coll: coll,
	}
	_, err = stor.ensureIndices(ctx)
	if err != nil {
		return nil, err
	}
	return stor, nil
}

func (s mongoStorage) Close() error {
	return s.conn.Disconnect(context.TODO())
}

func (s mongoStorage) ensureIndices(ctx context.Context) ([]string, error) {
	return s.coll.Indexes().CreateMany(ctx, indices)
}

func (s mongoStorage) Create(sub storage.Subscription) (string, error) {
	return "", nil // TODO
}

func (s mongoStorage) Read(id string) (storage.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (s mongoStorage) Update(id string, sub storage.Subscription) error {
	//TODO implement me
	panic("implement me")
}

func (s mongoStorage) Delete(id string) error {
	//TODO implement me
	panic("implement me")
}

func (s mongoStorage) List(limit uint32, cursor *string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (s mongoStorage) Resolve(limit uint32, cursor *string, _ []storage.PatternId) ([]string, error) {
	//TODO implement me
	panic("implement me")
}
