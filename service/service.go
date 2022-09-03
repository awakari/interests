package service

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/service/patterns"
	"github.com/meandros-messaging/subscriptions/storage"
	"github.com/meandros-messaging/subscriptions/util"
)

const (
	initialVersion = 0
)

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")
)

type (

	// NameCursor represents the Subscription name cursor.
	NameCursor *string

	// ResolveCursor represents the Subscription resolution cursors.
	ResolveCursor struct {

		// Name represents the last Subscription name.
		Name NameCursor

		// MetadataKey represents the last input metadata key used to resolve.
		MetadataKey *string

		// PatternId represents the last patterns.Code used to resolve.
		PatternId patterns.PatternCode
	}

	// Service is a Subscription CRUDL service.
	Service interface {

		// Create a Subscription means subscribing.
		// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
		Create(ctx context.Context, sub storage.Subscription) (err error)

		// Read the specified Subscription.
		// Returns ErrNotFound if Subscription is missing in the underlying storage.
		Read(ctx context.Context, name string) (sub storage.Subscription, err error)

		// Update replaces the existing Subscription.
		// Returns ErrNotFound if missing with the same Subscription.Name and Subscription.Version in the underlying storage.
		// In case of ErrNotFound it may be worth to Read the latest version again before Update retry.
		Update(ctx context.Context, sub storage.Subscription) (err error)

		// Delete a Subscription means unsubscribing.
		// Returns ErrNotFound if Subscription with the specified Subscription.Name is missing in the underlying storage.
		Delete(ctx context.Context, name string) (err error)

		// ListNames returns all known Subscription names with the pagination support.
		ListNames(ctx context.Context, limit uint32, cursor NameCursor) (page []string, err error)

		// ResolveNames returns all known Subscription names those matching the given message Metadata.
		ResolveNames(ctx context.Context, limit uint32, cursor ResolveCursor, md storage.Metadata) (page []string, err error)
	}

	service struct {
		patternsSvc       patterns.Service
		patternsPageLimit uint32
		s                 storage.Storage
	}
)

func NewService(patternsSvc patterns.Service, patternsPageLimit uint32, s storage.Storage) Service {
	return service{
		patternsSvc:       patternsSvc,
		patternsPageLimit: patternsPageLimit,
		s:                 s,
	}
}

func (svc service) Create(ctx context.Context, sub storage.Subscription) (err error) {
	return nil
}

func (svc service) Read(ctx context.Context, name string) (sub storage.Subscription, err error) {
	return storage.Subscription{}, nil
}

func (svc service) Update(ctx context.Context, sub storage.Subscription) (err error) {
	return nil
}

func (svc service) Delete(ctx context.Context, name string) (err error) {
	return nil
}

func (svc service) ListNames(ctx context.Context, limit uint32, cursor NameCursor) (page []string, err error) {
	return nil, nil
}

func (svc service) ResolveNames(ctx context.Context, limit uint32, cursor ResolveCursor, md storage.Metadata) (page []string, err error) {
	sortedKeys := util.SortedKeys(md)
	for _, k := range sortedKeys {
		if cursor.MetadataKey == nil || *cursor.MetadataKey <= k {
			svc.s.FindCandidates(ctx, limit-uint32(len(page)), cursor.Name, k, patternCode)
		}
	}
}
