package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/aggregator"
	"github.com/meandros-messaging/subscriptions/service/lexemes"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage"
	"golang.org/x/sync/errgroup"
)

type (
	// Service is a Subscription CRUDL service.
	Service interface {

		// Create an empty model.Subscription with the specified name and description.
		// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
		Create(ctx context.Context, name string, req CreateRequest) (err error)

		// Read the specified model.Subscription.
		// Returns ErrNotFound if Subscription is missing in the underlying storage.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Update applies the updates specified in the request to the existing model.Subscription.
		// Returns ErrNotFund f not exist or version mismatches.
		Update(ctx context.Context, subKey model.SubscriptionKey, req UpdateRequest) (err error)

		// Delete a model.Subscription and all associated model.Matcher those not in use by any other model.Subscription.
		// Returns ErrNotFound if model.Subscription with the specified name is missing in the underlying storage.
		Delete(ctx context.Context, name string) (err error)

		// ListNames returns all subscription names starting from the specified cursor.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// Resolve all matching subscriptions by the specified message model.Metadata and send them to aggregator.
		// Once returns all the matching subscriptions should be available in the aggregator with the specified
		// model.MessageId. It's client responsibility to filter the model.Subscription candidates from the aggregator.
		Resolve(ctx context.Context, md model.MessageDescriptor) (err error)
	}

	CreateRequest struct {
		Description string
		Includes    model.MatcherGroup
		Excludes    model.MatcherGroup
	}

	UpdateRequest struct {
		Description string
		Includes    MatcherGroupUpdate
		Excludes    MatcherGroupUpdate
	}

	MatcherGroupUpdate struct {
		All    bool
		Add    []model.Matcher
		Delete []model.Matcher
	}

	service struct {
		stor                     storage.Storage
		subsPageSizeLimit        uint32
		lexemesSvc               lexemes.Service
		excludesCompleteMatchers matchers.Service
		excludesPartialMatchers  matchers.Service
		includesCompleteMatchers matchers.Service
		includesPartialMatchers  matchers.Service
		matchersPageSizeLimit    uint32
		aggregatorSvc            aggregator.Service
	}
)

const (
	versionInitial = uint64(0)
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
)

func NewService(
	stor storage.Storage,
	subsPageSizeLimit uint32,
	lexemesSvc lexemes.Service,
	excludesCompleteMatchers matchers.Service,
	excludesPartialMatchers matchers.Service,
	includesCompleteMatchers matchers.Service,
	includesPartialMatchers matchers.Service,
	matchersPageSizeLimit uint32,
	aggregatorSvc aggregator.Service,
) Service {
	return service{
		stor:                     stor,
		subsPageSizeLimit:        subsPageSizeLimit,
		lexemesSvc:               lexemesSvc,
		excludesCompleteMatchers: excludesCompleteMatchers,
		excludesPartialMatchers:  excludesPartialMatchers,
		includesCompleteMatchers: includesCompleteMatchers,
		includesPartialMatchers:  includesPartialMatchers,
		matchersPageSizeLimit:    matchersPageSizeLimit,
		aggregatorSvc:            aggregatorSvc,
	}
}

func (svc service) Create(ctx context.Context, name string, req CreateRequest) (err error) {
	var sub model.Subscription
	sub.Name = name
	sub.Version = versionInitial
	sub.Description = req.Description
	sub.Includes.Matchers, err = svc.createMatchers(ctx, sub.Includes.Matchers, false)
	if err == nil {
		sub.Excludes.Matchers, err = svc.createMatchers(ctx, req.Excludes.Matchers, true)
		if err == nil {
			err = svc.stor.Create(ctx, sub)
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

func (svc service) Update(ctx context.Context, subKey model.SubscriptionKey, req UpdateRequest) (err error) {
	var sub model.Subscription
	sub, err = svc.stor.Read(ctx, subKey.Name)
	if err == nil {
		if sub.Version != subKey.Version {
			err = ErrNotFound
		} else {
			sub.Description = req.Description
			sub.Includes.All = req.Includes.All
			sub.Excludes.All = req.Excludes.All
			sub.Includes.Matchers = addAllMissing(req.Includes.Add, sub.Includes.Matchers)
			sub.Excludes.Matchers = addAllMissing(req.Excludes.Add, sub.Excludes.Matchers)
			sub.Includes.Matchers = delAllExisting(req.Includes.Delete, sub.Includes.Matchers)
			sub.Excludes.Matchers = delAllExisting(req.Excludes.Delete, sub.Excludes.Matchers)
			//
			g, gCtx := errgroup.WithContext(ctx)
			g.Go(func() error {
				_, createErr := svc.createMatchers(ctx, req.Includes.Add, false)
				return createErr
			})
			g.Go(func() error {
				_, createErr := svc.createMatchers(ctx, req.Excludes.Add, true)
				return createErr
			})
			g.Go(func() error {
				return svc.clearUnusedMatchers(ctx, req.Includes.Delete, false)
			})
			g.Go(func() error {
				return svc.clearUnusedMatchers(ctx, req.Excludes.Delete, true)
			})
			g.Go(func() error {
				return svc.stor.Update(gCtx, sub)
			})
			err = g.Wait()
		}
	}
	err = translateError(err)
	return
}

func addAllMissing(additions []model.Matcher, dst []model.Matcher) []model.Matcher {
	var exists bool
	for _, addition := range additions {
		exists = false
		for _, m := range dst {
			if addition.Equal(m) {
				exists = true
				break
			}
		}
		if !exists {
			dst = append(dst, addition)
		}
	}
	return dst
}

func delAllExisting(deletions []model.Matcher, dst []model.Matcher) (result []model.Matcher) {
	for _, m := range dst {
		for _, del := range deletions {
			if !del.Equal(m) {
				result = append(result, m)
			}
		}
	}
	return
}

func (svc service) Delete(ctx context.Context, name string) (err error) {
	var sub model.Subscription
	for {
		//
		sub, err = svc.stor.Read(ctx, name)
		if err != nil {
			break
		}
		// delete only if there's no change in the subscription matchers
		err = svc.stor.DeleteVersion(ctx, sub.SubscriptionKey)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				err = nil
				continue
			} else {
				break
			}
		}
		// delete matchers or decrement the corresponding reference counts also
		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return svc.clearUnusedMatchers(gCtx, sub.Includes.Matchers, false)
		})
		g.Go(func() error {
			return svc.clearUnusedMatchers(gCtx, sub.Excludes.Matchers, true)
		})
		err = g.Wait()
		break
	}
	err = translateError(err)
	return
}

func (svc service) clearUnusedMatchers(ctx context.Context, ms []model.Matcher, inExcludes bool) (err error) {
	for _, m := range ms {
		err = svc.deleteMatcherIfUnused(ctx, m, inExcludes)
		if err != nil {
			break
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
	err = svc.tryLockMatcher(ctx, m, inExcludes)
	if err == nil {
		defer svc.unlockMatcher(ctx, m, inExcludes)
		// find any subscription that is also using this matcher
		subs, err = svc.stor.Find(ctx, q, "")
		if err == nil {
			if len(subs) == 0 {
				// no other subscriptions found, let's delete the matcher from the corresponding storage
				err = svc.deleteMatcher(ctx, m, inExcludes)
			}
		}
	}
	return
}

func (svc service) tryLockMatcher(ctx context.Context, m model.Matcher, inExcludes bool) (err error) {
	matchersSvc := svc.selectMatchersService(inExcludes, m.Partial)
	return matchersSvc.TryLockCreate(ctx, m.MatcherData.Pattern.Code)
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

func (svc service) Resolve(ctx context.Context, md model.MessageDescriptor) (err error) {
	for k, v := range md.Metadata {
		err = svc.resolve(ctx, md.Id, k, v)
		if err != nil {
			break
		}
	}
	return
}

func (svc service) resolve(ctx context.Context, msgId model.MessageId, k, v string) (err error) {
	g, groupCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return svc.resolveMatchers(groupCtx, msgId, k, v, false, false)
	})
	g.Go(func() error {
		return svc.resolveMatchers(groupCtx, msgId, k, v, true, false)
	})
	valLexemes := svc.lexemesSvc.Split(v)
	for _, lexeme := range valLexemes {
		g.Go(func() error {
			return svc.resolveMatchers(groupCtx, msgId, k, lexeme, false, true)
		})
		g.Go(func() error {
			return svc.resolveMatchers(groupCtx, msgId, k, lexeme, true, true)
		})
	}
	return g.Wait()
}

func (svc service) resolveMatchers(
	ctx context.Context,
	msgId model.MessageId,
	k string,
	v string,
	inExcludes bool,
	isPartial bool,
) (
	err error,
) {
	g, groupCtx := errgroup.WithContext(ctx)
	var cursor model.PatternCode
	var page []model.PatternCode
	matchersSvc := svc.selectMatchersService(inExcludes, isPartial)
	for {
		page, err = matchersSvc.Search(ctx, k, v, svc.matchersPageSizeLimit, cursor)
		if err != nil {
			break
		}
		if len(page) == 0 {
			break
		}
		cursor = page[len(page)-1]
		for _, pc := range page {
			m := model.Matcher{
				MatcherData: model.MatcherData{
					Key: k,
					Pattern: model.Pattern{
						Code: pc,
					},
				},
				Partial: isPartial,
			}
			g.Go(func() error {
				return svc.resolveSubscriptions(groupCtx, msgId, m, inExcludes)
			})
		}
	}
	return g.Wait()
}

func (svc service) resolveSubscriptions(
	ctx context.Context,
	msgId model.MessageId,
	m model.Matcher,
	inExcludes bool,
) (
	err error,
) {
	g, groupCtx := errgroup.WithContext(ctx)
	var cursor string
	var page []model.Subscription
	q := storage.Query{
		Limit:      svc.subsPageSizeLimit,
		InExcludes: inExcludes,
		Matcher:    m,
	}
	for {
		page, err = svc.stor.Find(ctx, q, cursor)
		if err != nil {
			break
		}
		if len(page) == 0 {
			break
		}
		cursor = page[len(page)-1].Name
		for _, sub := range page {
			in := aggregator.MatchGroup{
				All:          sub.Includes.All,
				MatcherCount: uint32(len(sub.Includes.Matchers)),
			}
			ex := aggregator.MatchGroup{
				All:          sub.Excludes.All,
				MatcherCount: uint32(len(sub.Excludes.Matchers)),
			}
			match := aggregator.Match{
				MessageId:       msgId,
				SubscriptionKey: sub.SubscriptionKey,
				InExcludes:      inExcludes,
				Includes:        in,
				Excludes:        ex,
			}
			g.Go(func() error {
				return svc.aggregatorSvc.Update(groupCtx, match)
			})
		}
	}
	return g.Wait()
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
		case errors.Is(srcErr, matchers.ErrShouldRetry):
			dstErr = fmt.Errorf("%w: %s", ErrShouldRetry, srcErr)
		case errors.Is(srcErr, matchers.ErrInternal):
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		}
	}
	return
}
