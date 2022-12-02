package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage"
	"golang.org/x/sync/errgroup"
)

type (
	// Service is a model.Subscription CRUDL service.
	Service interface {

		// Create an empty model.Subscription with the specified name and description.
		// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
		// Returns model.ErrInvalidSubscription if the specified CreateRequest is invalid.
		Create(ctx context.Context, name string, req CreateRequest) (err error)

		// Read the specified model.Subscription.
		// Returns ErrNotFound if Subscription is missing in the underlying storage.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Delete a model.Subscription and all associated model.Matcher those not in use by any other model.Subscription.
		// Returns ErrNotFound if model.Subscription with the specified name is missing in the underlying storage.
		Delete(ctx context.Context, name string) (err error)

		// ListNames returns all subscription names starting from the specified cursor.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// Search returns subscriptions page where:<br/>
		// * model.Subscription name is greater than the one specified by the cursor<br/>
		// * subscriptions match the specified Query.
		Search(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error)
	}

	CreateRequest struct {
		Description string
		Routes      []string
		Includes    model.MatcherGroup
		Excludes    model.MatcherGroup
	}

	// Query represents the search query to use in Service.Search
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

	service struct {
		stor                     storage.Storage
		excludesCompleteMatchers matchers.Service
		excludesPartialMatchers  matchers.Service
		includesCompleteMatchers matchers.Service
		includesPartialMatchers  matchers.Service
	}
)

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")

	// ErrShouldRetry indicates a storage entity is locked and the operation should be retried.
	ErrShouldRetry = errors.New("retry the operation")

	// ErrInternal indicates some unexpected internal failure.
	ErrInternal = errors.New("internal failure")

	// ErrCleanMatcher indicates unused matchers cleanup failure upon a subscription deletion.
	ErrCleanMatcher = errors.New("matchers cleanup failure, may cause matchers garbage")
)

func NewService(
	stor storage.Storage,
	excludesCompleteMatchers matchers.Service,
	excludesPartialMatchers matchers.Service,
	includesCompleteMatchers matchers.Service,
	includesPartialMatchers matchers.Service,
) Service {
	return service{
		stor:                     stor,
		excludesCompleteMatchers: excludesCompleteMatchers,
		excludesPartialMatchers:  excludesPartialMatchers,
		includesCompleteMatchers: includesCompleteMatchers,
		includesPartialMatchers:  includesPartialMatchers,
	}
}

func (svc service) Create(ctx context.Context, name string, req CreateRequest) (err error) {
	var sub model.Subscription
	sub.Name = name
	sub.Description = req.Description
	sub.Routes = req.Routes
	sub.Includes = req.Includes
	sub.Excludes = req.Excludes
	err = sub.Validate()
	if err == nil {
		sub.Includes.Matchers, err = svc.createMatchers(ctx, req.Includes.Matchers, false)
		if err == nil {
			sub.Excludes.Matchers, err = svc.createMatchers(ctx, req.Excludes.Matchers, true)
			if err == nil {
				err = svc.stor.Create(ctx, sub)
			}
		}
	}
	err = translateError(err)
	return
}

func (svc service) createMatchers(
	ctx context.Context,
	matcherInputs []model.Matcher,
	inExcludes bool,
) (
	ms []model.Matcher,
	err error,
) {
	var md model.MatcherData
	for _, em := range matcherInputs {
		matchersSvc := svc.selectMatchersService(inExcludes, em.Partial)
		md, err = matchersSvc.Create(ctx, em.Key, em.Pattern.Src)
		if err != nil {
			break
		}
		m := model.Matcher{
			MatcherData: md,
			Partial:     em.Partial,
		}
		ms = append(ms, m)
	}
	return
}

func (svc service) selectMatchersService(inExcludes bool, partial bool) (matchersSvc matchers.Service) {
	if inExcludes {
		if partial {
			matchersSvc = svc.excludesPartialMatchers
		} else {
			matchersSvc = svc.excludesCompleteMatchers
		}
	} else {
		if partial {
			matchersSvc = svc.includesPartialMatchers
		} else {
			matchersSvc = svc.includesCompleteMatchers
		}
	}
	return
}

func (svc service) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	sub, err = svc.stor.Read(ctx, name)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Delete(ctx context.Context, name string) (err error) {
	var sub model.Subscription
	sub, err = svc.stor.Delete(ctx, name)
	if err == nil {
		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return svc.clearUnusedMatchers(gCtx, sub.Includes.Matchers, false)
		})
		g.Go(func() error {
			return svc.clearUnusedMatchers(gCtx, sub.Excludes.Matchers, true)
		})
		err = g.Wait()
		if err != nil {
			err = fmt.Errorf("%w: %s, subscription: %v", ErrCleanMatcher, err, sub)
		}
	}
	err = translateError(err)
	return
}

func (svc service) clearUnusedMatchers(ctx context.Context, ms []model.Matcher, inExcludes bool) (firstErr error) {
	for _, m := range ms {
		err := svc.deleteMatcherIfUnused(ctx, m, inExcludes)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return
}

func (svc service) deleteMatcherIfUnused(ctx context.Context, m model.Matcher, inExcludes bool) (err error) {
	q := storage.Query{
		Limit:      1,
		InExcludes: inExcludes,
		Matcher:    m,
	}
	var subs []model.Subscription
	err = svc.lockMatcher(ctx, m, inExcludes)
	if err == nil {
		defer svc.unlockMatcher(ctx, m, inExcludes)
		// find any subscription that is also using this matcher
		subs, err = svc.stor.Search(ctx, q, "")
		if err == nil {
			if len(subs) == 0 {
				// no other subscriptions found, let's delete the matcher from the corresponding storage
				err = svc.deleteMatcher(ctx, m, inExcludes)
			}
		}
	}
	return
}

func (svc service) lockMatcher(ctx context.Context, m model.Matcher, inExcludes bool) error {
	matchersSvc := svc.selectMatchersService(inExcludes, m.Partial)
	return matchersSvc.LockCreate(ctx, m.MatcherData.Pattern.Code)
}

func (svc service) unlockMatcher(ctx context.Context, m model.Matcher, inExcludes bool) {
	matchersSvc := svc.selectMatchersService(inExcludes, m.Partial)
	_ = matchersSvc.UnlockCreate(ctx, m.MatcherData.Pattern.Code)
}

func (svc service) deleteMatcher(ctx context.Context, m model.Matcher, inExcludes bool) (err error) {
	matchersSvc := svc.selectMatchersService(inExcludes, m.Partial)
	return matchersSvc.Delete(ctx, m.MatcherData)
}

func (svc service) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	page, err = svc.stor.ListNames(ctx, limit, cursor)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Search(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error) {
	storageQuery := storage.Query{
		Limit:      q.Limit,
		InExcludes: q.InExcludes,
		Matcher:    q.Matcher,
	}
	page, err = svc.stor.Search(ctx, storageQuery, cursor)
	if err != nil {
		err = translateError(err)
	}
	return
}

func translateError(srcErr error) (dstErr error) {
	if srcErr == nil {
		dstErr = nil
	} else {
		switch {
		case errors.Is(srcErr, storage.ErrConflict):
			dstErr = fmt.Errorf("%w: %s", ErrConflict, srcErr)
		case errors.Is(srcErr, storage.ErrNotFound):
			dstErr = fmt.Errorf("%w: %s", ErrNotFound, srcErr)
		case errors.Is(srcErr, storage.ErrInternal):
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		case errors.Is(srcErr, matchers.ErrShouldRetry):
			dstErr = fmt.Errorf("%w: %s", ErrShouldRetry, srcErr)
		case errors.Is(srcErr, matchers.ErrInternal):
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		case errors.Is(srcErr, model.ErrInvalidSubscription):
			dstErr = srcErr
		case errors.Is(srcErr, ErrNotFound):
			dstErr = srcErr
		case errors.Is(srcErr, ErrInternal):
			dstErr = srcErr
		case errors.Is(srcErr, ErrConflict):
			dstErr = srcErr
		case errors.Is(srcErr, ErrShouldRetry):
			dstErr = srcErr
		case errors.Is(srcErr, ErrCleanMatcher):
			dstErr = srcErr
		default:
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		}
	}
	return
}
