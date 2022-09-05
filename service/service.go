package service

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/patterns"
	"github.com/meandros-messaging/subscriptions/storage"
	"sort"
)

const (
	initialVersion    = 0
	patternsPageLimit = 1000
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

	// ResolveCursor represents the model.Subscription resolution cursors.
	ResolveCursor struct {

		// Name represents the last Subscription name.
		Name NameCursor

		// PatternCode represents the last patterns.Code used to resolve.
		PatternCode model.PatternCode
	}

	// ResolveResult represents the model.Subscription page together with next results page cursor.
	ResolveResult struct {

		// NextPageCursor is the cursor to resolve the next page of results.
		NextPageCursor ResolveCursor

		// MetadataKey represents the last input metadata key used to resolve.
		MetadataKey string

		// NamesPage contains the model.Subscription names.
		NamesPage []string
	}

	// Service is a Subscription CRUDL service.
	Service interface {

		// Create a Subscription means subscribing.
		// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
		Create(ctx context.Context, sub model.Subscription) (err error)

		// Read the specified Subscription.
		// Returns ErrNotFound if Subscription is missing in the underlying storage.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Update replaces the existing Subscription.
		// Returns ErrNotFound if missing with the same Subscription.Name and Subscription.Version in the underlying storage.
		// In case of ErrNotFound it may be worth to Read the latest version again before Update retry.
		Update(ctx context.Context, sub model.Subscription) (err error)

		// Delete a Subscription means unsubscribing.
		// Returns ErrNotFound if Subscription with the specified Subscription.Name is missing in the underlying storage.
		Delete(ctx context.Context, name string) (err error)

		// ListNames returns all known model.Subscription names with the pagination support.
		ListNames(ctx context.Context, limit uint32, cursor NameCursor) (page []string, err error)

		// ResolveNames returns all known model.Subscription names those matching the given message model.Metadata.
		// Note that it's a caller responsibility to remove the processed keys from the input model.Metadata.
		// So to resolve a next page, remove from the model.Metadata all keys that are less than ResolveResult.MetadataKey.
		ResolveNames(ctx context.Context, limit uint32, cursor ResolveCursor, md model.Metadata) (r ResolveResult, err error)
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

func (svc service) Create(ctx context.Context, sub model.Subscription) (err error) {
	return nil
}

func (svc service) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	return model.Subscription{}, nil
}

func (svc service) Update(ctx context.Context, sub model.Subscription) (err error) {
	return nil
}

func (svc service) Delete(ctx context.Context, name string) (err error) {
	return nil
}

func (svc service) ListNames(ctx context.Context, limit uint32, cursor NameCursor) (page []string, err error) {
	return nil, nil
}

func (svc service) ResolveNames(ctx context.Context, limit uint32, cursor ResolveCursor, md model.Metadata) (r ResolveResult, err error) {
	if len(md) > 0 {
		sortedMdKeys := make([]string, 0, len(md))
		for k, _ := range md {
			sortedMdKeys = append(sortedMdKeys, k)
		}
		sort.Strings(sortedMdKeys)
		mdKeyCursor := sortedMdKeys[0]
		r, err = svc.resolveNames(ctx, limit, cursor, md, mdKeyCursor)
	}
	return
}

func (svc service) resolveNames(ctx context.Context, limit uint32, cursor ResolveCursor, md model.Metadata, mdKeyCursor string) (r ResolveResult, err error) {
	var patternCodesByMdKey map[string][]model.PatternCode
	r.MetadataKey = mdKeyCursor
	var subs []model.Subscription
	for { // loop over pattern code pages
		remainingLimit := limit - uint32(len(r.NamesPage))
		if remainingLimit > 0 {
			patternsCursor := patterns.BulkCursor{
				Key:         r.MetadataKey,
				PatternCode: cursor.PatternCode,
			}
			patternCodesByMdKey, err = svc.patternsSvc.SearchMatchesBulk(ctx, md, patternsPageLimit, &patternsCursor)
			patternCodesByMdKeyLen := len(patternCodesByMdKey)
			if err == nil && patternCodesByMdKeyLen > 0 {
				sortedMdKeys := make([]string, 0, patternCodesByMdKeyLen)
				for k, _ := range patternCodesByMdKey {
					sortedMdKeys = append(sortedMdKeys, k)
				}
				sort.Strings(sortedMdKeys)
				for _, k := range sortedMdKeys {
					remainingLimit = limit - uint32(len(r.NamesPage))
					if remainingLimit > 0 {
						r.MetadataKey = k
						for _, patternCode := range patternCodesByMdKey[k] {
							r.NextPageCursor.PatternCode = patternCode
							err = svc.getMatchingSubscriptions(ctx, limit, md, k, patternCode, &r)
							if err != nil {
								break
							}
						}
					} else {
						break
					}
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

func (svc service) getMatchingSubscriptions(ctx context.Context, limit uint32, md model.Metadata, k string, patternCode model.PatternCode, r *ResolveResult) (err error) {
	var subs []model.Subscription
	for { // loop over subscription pages
		remainingLimit := limit - uint32(len(r.NamesPage))
		if remainingLimit > 0 {
			subs, err = svc.s.FindCandidates(ctx, remainingLimit, r.NextPageCursor.Name, k, patternCode)
			if err == nil && len(subs) > 0 {
				for _, sub := range subs {
					if sub.Matches(md, k, patternCode) {
						r.NamesPage = append(r.NamesPage, sub.Name)
					}
				}
				r.NextPageCursor.Name = &subs[len(subs)-1].Name
			} else {
				break
			}
		} else {
			break
		}
	}
	return
}
