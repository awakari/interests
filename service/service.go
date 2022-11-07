package service

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service/aggregator"
	"github.com/meandros-messaging/subscriptions/service/lexemes"
	"github.com/meandros-messaging/subscriptions/service/matchers"
	"github.com/meandros-messaging/subscriptions/storage"
	"golang.org/x/sync/errgroup"
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
			break
		}
		m := model.Matcher{
			MatcherData: md,
			Partial:     em.Partial,
		}
		ms = append(ms, m)
	}
	if err != nil {
		err = translateError(err)
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
			break
		}
		m := model.Matcher{
			MatcherData: md,
			Partial:     em.Partial,
		}
		ms = append(ms, m)
	}
	if err != nil {
		err = translateError(err)
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
		// 1. Read the subscription version and matchers
		sub, err = svc.Read(ctx, name)
		if err != nil {
			break
		}
		// 2. For every subscription matcher check if there's any other subscription using it before removing
		for _, m := range sub.Excludes.Matchers {
			subs, err = svc.stor.FindByMatcherData(ctx, 1, "", m.MatcherData)
			if err != nil {
				break
			}
			// FIXME unsafe, matcher may become used after the check and actual deletion
			if len(subs) == 0 {
				if m.Partial {
					err = svc.excludesPartialMatchers.Delete(ctx, m.MatcherData)
				} else {
					err = svc.excludesCompleteMatchers.Delete(ctx, m.MatcherData)
				}
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			err = translateError(err)
			break
		}
		//
		for _, m := range sub.Includes.Matchers {
			subs, err = svc.stor.FindByMatcherData(ctx, 1, "", m.MatcherData)
			if err != nil {
				break
			}
			if len(subs) == 0 {
				if m.Partial {
					err = svc.includesPartialMatchers.Delete(ctx, m.MatcherData)
				} else {
					err = svc.includesCompleteMatchers.Delete(ctx, m.MatcherData)
				}
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			err = translateError(err)
			break
		}
		//
		err = svc.stor.DeleteVersion(ctx, sub.SubscriptionKey)
		if errors.Is(err, storage.ErrNotFound) {
			err = nil
		} else {
			err = translateError(err)
			break
		}
	}
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
		return svc.resolveMatchers(groupCtx, msgId, k, v, svc.excludesCompleteMatchers)
	})
	g.Go(func() error {
		return svc.resolveMatchers(groupCtx, msgId, k, v, svc.includesCompleteMatchers)
	})
	valLexemes := svc.lexemesSvc.Split(v)
	for _, lexeme := range valLexemes {
		g.Go(func() error {
			return svc.resolveMatchers(groupCtx, msgId, k, lexeme, svc.excludesPartialMatchers)
		})
		g.Go(func() error {
			return svc.resolveMatchers(groupCtx, msgId, k, lexeme, svc.includesPartialMatchers)
		})
	}
	return g.Wait()
}

func (svc service) resolveMatchers(ctx context.Context, msgId model.MessageId, k, v string, matchersSvc matchers.Service) (err error) {
	g, groupCtx := errgroup.WithContext(ctx)
	var cursor model.PatternCode
	var page []model.PatternCode
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
			md := model.MatcherData{
				Key: k,
				Pattern: model.Pattern{
					Code: pc,
				},
			}
			g.Go(func() error {
				return svc.resolveSubscriptions(groupCtx, msgId, md)
			})
		}
	}
	return g.Wait()
}

func (svc service) resolveSubscriptions(ctx context.Context, msgId model.MessageId, md model.MatcherData) (err error) {
	g, groupCtx := errgroup.WithContext(ctx)
	var cursor string
	var page []model.Subscription
	for {
		page, err = svc.stor.FindByMatcherData(ctx, svc.subsPageSizeLimit, cursor, md)
		if err != nil {
			break
		}
		if len(page) == 0 {
			break
		}
		cursor = page[len(page)-1].Name
		for _, sub := range page {
			in := aggregator.MatchInGroup{
				All:          sub.Includes.All,
				MatcherCount: uint32(len(sub.Includes.Matchers)),
			}
			ex := aggregator.MatchInGroup{
				All:          sub.Excludes.All,
				MatcherCount: uint32(len(sub.Excludes.Matchers)),
			}
			m := aggregator.Match{
				MessageId:       msgId,
				SubscriptionKey: sub.SubscriptionKey,
				Includes:        &in,
				Excludes:        &ex,
			}
			g.Go(func() error {
				return svc.aggregatorSvc.Update(groupCtx, m)
			})
		}
	}
	return g.Wait()
}

func translateError(srcErr error) (dstErr error) {
	if srcErr == nil {
		dstErr = nil
	} else {
		// TODO
		dstErr = srcErr
	}
	return
}
