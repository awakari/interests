package storage

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/model/subscription"
	"io"
)

type (

	// Storage represents the subscriptions storage
	Storage interface {
		io.Closer

		// Create a subscription with the specified account and data.
		// Returns a created subscription id if successful.
		Create(ctx context.Context, acc string, sd subscription.Data) (id string, err error)

		// Read the subscription.Data by the subscription.Subscription id.
		Read(ctx context.Context, id, acc string) (sd subscription.Data, err error)

		// UpdateMetadata updates the mutable part of the subscription.Data
		UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error)

		// Delete removes the subscription.Subscription specified by its unique id.
		// Returns the subscription.Data if deleted, error otherwise.
		Delete(ctx context.Context, id, acc string) (sd subscription.Data, err error)

		// SearchByAccount returns all subscription ids those have the account matching the query.
		SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error)

		// SearchByKiwi returns subscriptions page where:<br/>
		// * model.Subscription id is greater than the one specified by the cursor<br/>
		// * subscriptions match the specified model.KiwiQuery.
		SearchByKiwi(ctx context.Context, q KiwiQuery, cursor string) (page []subscription.ConditionMatch, err error)
	}
)

var (

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")

	// ErrInternal indicates the internal storage failure happened.
	ErrInternal = errors.New("internal subscription storage failure")
)
