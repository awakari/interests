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
		Create(ctx context.Context, sub model.Subscription) (err error)

		// Read the specified subscription details.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Delete removes the model.Subscription specified by its unique name.
		// Returns the model.Subscription if deleted, error otherwise.
		Delete(ctx context.Context, name string) (sub model.Subscription, err error)

		// ListNames returns all known subscription names with the pagination support that match the specified query.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// SearchByKiwi returns subscriptions page where:<br/>
		// * model.Subscription name is greater than the one specified by the cursor<br/>
		// * subscriptions match the specified model.KiwiQuery.
		SearchByKiwi(ctx context.Context, q KiwiQuery, cursor string) (page []model.Subscription, err error)
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
