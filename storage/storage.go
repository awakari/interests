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
		Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error)

		// Read the subscription.Data by the subscription.Subscription id.
		Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error)

		// Update updates the subscription.Data
		Update(ctx context.Context, id, groupId, userId string, sd subscription.Data) (prev subscription.Data, err error)

		// UpdateFollowers updates the followers count
		UpdateFollowers(ctx context.Context, id string, count int64) (err error)

		// Delete removes the subscription.Subscription specified by its unique id.
		// Returns the subscription.Data if deleted, error otherwise.
		Delete(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error)

		// Search returns all subscription ids matching the query.
		Search(ctx context.Context, q subscription.Query, cursor subscription.Cursor) (ids []string, err error)

		// SearchByCondition finds all subscriptions those match the specified condition id and feeds these to the
		// specified consumer func.
		SearchByCondition(ctx context.Context, q subscription.QueryByCondition, cursor string) (page []subscription.ConditionMatch, err error)

		Count(ctx context.Context) (count int64, err error)
		CountUsersUnique(ctx context.Context) (count int64, err error)
	}
)

var (

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")

	// ErrInternal indicates the internal storage failure happened.
	ErrInternal = errors.New("internal subscription storage failure")
)
