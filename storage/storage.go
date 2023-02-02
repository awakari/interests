package storage

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/model"
	"io"
)

type (

	// Storage represents the subscriptions storage
	Storage interface {
		io.Closer

		// Create a subscription means subscribing.
		Create(ctx context.Context, sub model.SubscriptionData) (id string, err error)

		// Read the model.SubscriptionData by the model.Subscription id.
		Read(ctx context.Context, id string) (sub model.SubscriptionData, err error)

		// Delete removes the model.Subscription specified by its unique id.
		// Returns the model.SubscriptionData if deleted, error otherwise.
		Delete(ctx context.Context, name string) (sub model.SubscriptionData, err error)

		// SearchByKiwi returns subscriptions page where:<br/>
		// * model.Subscription id is greater than the one specified by the cursor<br/>
		// * subscriptions match the specified model.KiwiQuery.
		SearchByKiwi(ctx context.Context, q KiwiQuery, cursor string) (page []model.Subscription, err error)

		// SearchByMetadata returns all subscriptions those have the metadata matching the query (same keys and values).
		SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []model.Subscription, err error)
	}
)

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")

	// ErrInternal indicates the internal storage failure happened.
	ErrInternal = errors.New("internal subscription storage failure")
)
