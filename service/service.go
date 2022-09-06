package service

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/patterns"
	"github.com/meandros-messaging/subscriptions/storage"
	"github.com/meandros-messaging/subscriptions/util"
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

		// Patterns represents the last model.Pattern page.
		Patterns patterns.BulkCursor

		// Name represents the last model.Subscription name.
		Name NameCursor
	}

	// ResolveResult represents the model.Subscription page together with next results page cursor.
	ResolveResult struct {

		// NextPageCursor is the cursor to resolve the next page of results.
		NextPageCursor ResolveCursor

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
		ResolveNames(ctx context.Context, limit uint32, cursor *ResolveCursor, md model.Metadata) (r ResolveResult, err error)
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

func (svc service) ResolveNames(ctx context.Context, limit uint32, cursor *ResolveCursor, md model.Metadata) (r ResolveResult, err error) {
	if len(md) > 0 {
		r, err = svc.resolveNames(ctx, limit, cursor, md)
	}
	return
}

func (svc service) resolveNames(ctx context.Context, limit uint32, cursor *ResolveCursor, md model.Metadata) (r ResolveResult, err error) {
	var patternCodesByMdKey map[string][]model.PatternCode
	if cursor != nil {
		r.NextPageCursor.Name = cursor.Name
		patternsCursor := patterns.BulkCursor{
			Key:         cursor.Patterns.Key,
			PatternCode: cursor.Patterns.PatternCode,
		}
		r.NextPageCursor.Patterns = patternsCursor
	}
	for { // loop over pattern code pages
		remainingLimit := limit - uint32(len(r.NamesPage))
		if remainingLimit > 0 {
			patternCodesByMdKey, err = svc.patternsSvc.SearchMatchesBulk(ctx, md, patternsPageLimit, &cursor.Patterns)
			if err == nil && len(patternCodesByMdKey) > 0 {
				sortedMdKeys := util.SortedKeys(patternCodesByMdKey)
				for _, k := range sortedMdKeys {
					remainingLimit = limit - uint32(len(r.NamesPage))
					if remainingLimit > 0 {
						r.NextPageCursor.Patterns.Key = k
						for _, patternCode := range patternCodesByMdKey[k] {
							r.NextPageCursor.Patterns.PatternCode = patternCode
							err = svc.findAndSetToResult(ctx, limit, md, k, patternCode, &r)
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

func (svc service) findAndSetToResult(ctx context.Context, limit uint32, md model.Metadata, k string, patternCode model.PatternCode, r *ResolveResult) (err error) {
	var subs []model.Subscription
	for { // loop over subscription pages
		remainingLimit := limit - uint32(len(r.NamesPage))
		if remainingLimit > 0 {
			subs, err = svc.s.FindCandidates(ctx, remainingLimit, r.NextPageCursor.Name, k, patternCode)
			if err == nil && len(subs) > 0 {
				r.filterMatching(subs, md, k, patternCode)
			} else {
				break
			}
		} else {
			break
		}
	}
	return
}

func (r *ResolveResult) filterMatching(subs []model.Subscription, md model.Metadata, k string, patternCode model.PatternCode) {
	for _, sub := range subs {
		if sub.Matches(md, k, patternCode) {
			r.NamesPage = append(r.NamesPage, sub.Name)
		}
	}
	r.NextPageCursor.Name = &subs[len(subs)-1].Name
}
