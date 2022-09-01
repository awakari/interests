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

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")
)

type (

	// SortedMetadata is the incoming message metadata to match the subscriptions.
	// Note that keys should be sorted.
	SortedMetadata map[string]string

	// MetadataPatternIds is the map of the matched patterns by SortedMetadata keys
	MetadataPatternIds map[string][]patterns.Code

	// NameCursor represents the Subscription name cursor.
	NameCursor *string

	// ResolveCursor represents the Subscription resolution cursors.
	ResolveCursor struct {

		// Name represents the last Subscription name.
		Name NameCursor

		// MetadataKey represents the last input metadata key used to resolve.
		MetadataKey *string

		// PatternId represents the last patterns.Code used to resolve.
		PatternId patterns.Code
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
		List(ctx context.Context, limit uint32, cursor NameCursor) ([]string, error)

		// Resolve returns all known Subscription names those matching the given message metadata.
		Resolve(ctx context.Context, limit uint32, cursor ResolveCursor, md SortedMetadata) ([]string, error)
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

func (svc service) List(ctx context.Context, limit uint32, cursor NameCursor) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (svc service) Resolve(ctx context.Context, limit uint32, cursor ResolveCursor, md SortedMetadata) (names []string, err error) {
	keyCursor := cursor.MetadataKey
	patternCursor := cursor.PatternId
	nameCursor := cursor.Name
	var patternCodes []patterns.Code
	var nextNames []string
	for k, v := range md {
		if keyCursor != nil && *keyCursor < k {
			patternCodes, nextNames, err = svc.resolvePatterns(ctx, limit, k, v, patternCursor, nameCursor)
			if err == nil {
				if len(nextNames) > 0 {
					nameCursor = &nextNames[len(nextNames)-1]
					names = append(names, nextNames...)
				}
				if len(patternCodes) > 0 {
					patternCursor = patternCodes[len(patternCodes)-1]
				}
			} else {
				break
			}
		}
	}
	return
}

func (svc service) resolvePatterns(ctx context.Context, limit uint32, mdKey string, mdVal string, patternCursor patterns.Code, nameCursor NameCursor) (patternCodes []patterns.Code, names []string, err error) {
	var nextNames []string
	for {
		patternCodes, err = svc.patternsSvc.SearchMatches(ctx, mdVal, svc.patternsPageLimit, patternCursor)
		if err == nil {
			if len(patternCodes) > 0 {
				patternCursor = patternCodes[len(patternCodes)-1]
				nextNames, err = svc.resolveNames(ctx, limit-uint32(len(names)), mdKey, patternCodes, nameCursor)
				if err == nil {
					if len(nextNames) > 0 {
						nameCursor = &nextNames[len(nextNames)-1]
						names = append(names, nextNames...)
					}
				} else {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}
	return
}

func (svc service) resolveNames(ctx context.Context, limit uint32, mdKey string, patternCodes []patterns.Code, cursor NameCursor) (names []string, err error) {
	var page []string
	for {
		page, err = svc.s.Resolve(ctx, limit-uint32(len(names)), cursor, mdKey, patternCodes)
		if err == nil {
			if len(names) > 0 {
				cursor = &page[len(page)-1]
				names = append(names, page...)
			} else {
				break
			}
		} else {
			break
		}
	}
	return
}
