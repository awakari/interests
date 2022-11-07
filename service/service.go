package service

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/aggregator"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage"
)

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")
)

type (

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

		// ListNames returns all subscription names starting from the specified cursor.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// Resolve all matching subscriptions by the specified message model.Metadata and send them to aggregator.
		// Once returns all the matching subscriptions should be available in the aggregator with the specified
		// model.MessageId. It's client responsibility to filter the model.Subscription candidates from the aggregator.
		Resolve(ctx context.Context, md model.MessageDescriptor) (err error)
	}

	service struct {
		stor                     storage.Storage
		excludesCompleteMatchers matchers.Service
		excludesPartialMatchers  matchers.Service
		includesCompleteMatchers matchers.Service
		includesPartialMatchers  matchers.Service
		matchersPageSizeLimit    uint32
		aggregatorSvc            aggregator.Service
	}
)

func NewService(
	stor storage.Storage,
	excludesCompleteMatchers matchers.Service,
	excludesPartialMatchers matchers.Service,
	includesCompleteMatchers matchers.Service,
	includesPartialMatchers matchers.Service,
	matchersPageSizeLimit uint32,
	aggregatorSvc aggregator.Service,
) Service {
	return service{
		stor:                     stor,
		excludesCompleteMatchers: excludesCompleteMatchers,
		excludesPartialMatchers:  excludesPartialMatchers,
		includesCompleteMatchers: includesCompleteMatchers,
		includesPartialMatchers:  includesPartialMatchers,
		matchersPageSizeLimit:    matchersPageSizeLimit,
		aggregatorSvc:            aggregatorSvc,
	}
}

func (svc service) Create(ctx context.Context, sub model.Subscription) (err error) {
	//
	var ms []model.Matcher
	var md model.MatcherData
	//
	for _, em := range sub.Excludes.Matchers {
		if em.Partial {
			md, err = svc.excludesPartialMatchers.Create(ctx, em.Key, em.Pattern.Src)
		} else {
			md, err = svc.excludesCompleteMatchers.Create(ctx, em.Key, em.Pattern.Src)
		}
		if err != nil {
			err = translateError(err)
			break
		}
		m := model.Matcher{
			MatcherData: md,
			Partial:     em.Partial,
		}
		ms = append(ms, m)
	}
	if err != nil {
		return
	}
	sub.Excludes.Matchers = ms
	//
	for _, em := range sub.Includes.Matchers {
		if em.Partial {
			md, err = svc.includesPartialMatchers.Create(ctx, em.Key, em.Pattern.Src)
		} else {
			md, err = svc.includesCompleteMatchers.Create(ctx, em.Key, em.Pattern.Src)
		}
		if err != nil {
			err = translateError(err)
			break
		}
		m := model.Matcher{
			MatcherData: md,
			Partial:     em.Partial,
		}
		ms = append(ms, m)
	}
	if err != nil {
		return
	}
	sub.Includes.Matchers = ms
	//
	err = svc.stor.Create(ctx, sub)
	if err != nil {
		err = translateError(err)
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

func (svc service) Update(ctx context.Context, sub model.Subscription) (err error) {
	err = svc.stor.DeleteVersion(ctx, sub.SubscriptionKey)
	if err != nil {
		err = translateError(err)
	} else {
		err = svc.Create(ctx, sub)
	}
	return
}

func (svc service) Delete(ctx context.Context, name string) (err error) {
	var sub model.Subscription
	var subs []model.Subscription
	for {
		sub, err = svc.Read(ctx, name)
		if err != nil {
			break
		}
		for _, m := range sub.Excludes.Matchers {
			subs, err = svc.stor.FindByMatcherData(ctx, 1, "", m.MatcherData)
			if err != nil {
				err = translateError(err)
				break
			}
			if len(subs) == 0 {
				svc.matchersSvc.Delete(ctx)
			}
		}
	}

	// 1. Read
	// 2. For every subscription matcher check if there's any other subscription using it, if none - remove from matchers
	// 3. invoke DeleteVersion in storage
	return
}

func (svc service) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	page, err = svc.stor.ListNames(ctx, limit, cursor)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Resolve(ctx context.Context, md model.MessageDescriptor) (err error) {
	var patternCodesPage []model.PatternCode
	var patternCodeCursor model.PatternCode
	for k, v := range md.Metadata {
		for {
			patternCodesPage, err = svc.matchersSvc.Search(ctx, k, v, svc.matchersPageSizeLimit, patternCodeCursor)
			if err != nil {
				err = translateError(err)
				break
			}
			if len(patternCodesPage) == 0 {
				break
			}
			patternCodeCursor = patternCodesPage[len(patternCodesPage)-1]
			for _, patternCode := range patternCodesPage {
				matcherData := model.MatcherData{
					Key: k,
					Pattern: model.Pattern{
						Code: patternCode,
					},
				}

			}
		}
	}
	return
}

func (svc service) createMatchers(ctx context.Context, srcMatchers []model.Matcher) (dstMatchers []model.Matcher, err error) {
	var md model.MatcherData
	for _, em := range srcMatchers {
		md, err = svc.matchersSvc.Create(ctx, em.Key, em.Pattern.Src)
		if err != nil {
			err = translateError(err)
			break
		}
		m := model.Matcher{
			MatcherData: md,
			Partial:     em.Partial,
		}
		dstMatchers = append(dstMatchers, m)
	}
	return
}

func translateError(srcErr error) (dstErr error) {
	// TODO
	dstErr = srcErr
	return
}
