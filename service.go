package subscriptions

import (
	"context"
	"errors"
	"subscriptions/patterns"
	"subscriptions/storage"
)

const (
	initialVersion = 0
)

var (
	ErrConflict = errors.New("subscription already exists")

	ErrNotFound = errors.New("subscription was not found")
)

type (

	// SortedMetadata is the incoming message metadata to match the subscriptions.
	// Note that keys should be sorted.
	SortedMetadata map[string]string

	// MetadataPatternIds is the map of the matched patterns by SortedMetadata keys
	MetadataPatternIds map[string][]patterns.Id

	Cursor struct {
		MetadataKey *string
		PatternId   patterns.Id
		Name        *string
	}

	// Service is a subscriptions service.
	Service interface {

		// Create a subscription means subscribing.
		// Returns ErrConflict if a subscriptions with the same name already present in the underlying storage.
		Create(ctx context.Context, data Data) error

		// Read the specified subscription details.
		// Returns ErrNotFound if subscription is missing in the underlying storage.
		Read(ctx context.Context, name string) (*Data, error)

		// Update creates or updates the existing subscription with the specified subscription Data.
		// Returns ErrNotFound if subscription is missing with the same name and version in the underlying storage.
		// In case of ErrNotFound it may be worth to Read the latest version again before Update retry.
		Update(ctx context.Context, version uint64, data Data) error

		// Delete a subscription means unsubscribing.
		// Returns ErrNotFound if subscription is missing in the underlying storage.
		Delete(ctx context.Context, name string) error

		// List returns all known Subscription names with the pagination support.
		List(ctx context.Context, limit uint32, cursor *string) ([]string, error)

		// Resolve returns all known Subscription names those matching the given message metadata.
		Resolve(ctx context.Context, limit uint32, cursor Cursor, md SortedMetadata) ([]string, error)
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

func (svc service) Create(ctx context.Context, data Data) (err error) {
	return svc.Update(ctx, initialVersion, data)
}

func (svc service) Read(ctx context.Context, name string) (*Data, error) {
	return nil, nil // TODO)
}

func (svc service) Update(ctx context.Context, version uint64, data Data) error {
	//TODO implement me
	panic("implement me")
}

func (svc service) Delete(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}

func (svc service) List(ctx context.Context, limit uint32, cursor *string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (svc service) Resolve(ctx context.Context, limit uint32, cursor Cursor, md SortedMetadata) (names []string, err error) {
	keyCursor := cursor.MetadataKey
	patternCursor := cursor.PatternId
	nameCursor := cursor.Name
	var namesPage []string
	for k, v := range md {
		if keyCursor != nil && *keyCursor < k {
			var patternIds []patterns.Id
			for {
				patternIds, err = svc.patternsSvc.SearchMatches(ctx, v, svc.patternsPageLimit, patternCursor)
				if err == nil {
					if len(patternIds) > 0 {
						patternCursor = patternIds[len(patternIds)-1]
						for {
							namesPage, err = svc.s.Resolve(limit-uint32(len(names)), nameCursor, patternIds)
							if err == nil {
								if len(names) > 0 {
									nameCursor = &namesPage[len(namesPage)-1]
								}
							}
						}
					} else {
						break
					}
				}
			}
		}
	}
	return
}
