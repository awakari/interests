package storage

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"io"
)

type (

	// Query represents the search query to use in Storage.Find
	Query struct {

		// Limit defines a results page size limit.
		Limit uint32

		// InExcludes defines if it's necessary to find a model.Subscription with same model.Matcher in the "InExcludes"
		// model.MatcherGroup
		InExcludes bool

		// Matcher represents a model.Matcher that should be present in the model.Subscription to include into the
		// search results.
		Matcher model.Matcher
	}

	// Storage represents the subscriptions storage
	Storage interface {
		io.Closer

		// Create a subscription means subscribing.
		Create(ctx context.Context, sub model.Subscription) (err error)

		// Read the specified subscription details.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Delete removes the model.Subscription specified by its unique name.
		Delete(ctx context.Context, name string) (err error)

		// ListNames returns all known subscription names with the pagination support that match the specified query.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// Find returns subscriptions page where:<br/>
		// * model.Subscription name is greater than the one specified by the cursor<br/>
		// * subscriptions match the specified Query.
		Find(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error)
	}
)

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")
)
