package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/api/grpc/kiwi-tree"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"github.com/awakari/subscriptions/util"
	"reflect"
)

// Service is a subscription.Subscription CRUDL service.
type Service interface {

	// Create a new subscription.Subscription with the specified fields.
	// Returns subscription.ErrInvalidSubscriptionCondition if the specified condition.Condition is invalid.
	Create(ctx context.Context, acc string, d subscription.Data) (id string, err error)

	// Read the specified subscription.Subscription.
	// Returns ErrNotFound if subscription.Subscription is missing in the underlying storage.
	Read(ctx context.Context, id, acc string) (d subscription.Data, err error)

	// UpdateMetadata updates the mutable part of the subscription.Data
	UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error)

	// Delete a subscription.Subscription and all associated conditions those not in use by any other subscription.
	// Returns ErrNotFound if a subscription.Subscription with the specified id is missing in the underlying storage.
	Delete(ctx context.Context, id, acc string) (err error)

	// SearchByAccount returns all subscription ids those have the account matching the query.
	SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error)

	// SearchByCondition sends the matches to the specified consumer func those contain a condition specified by the query.
	SearchByCondition(ctx context.Context, cond condition.Condition, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error)
}

type service struct {
	stor                storage.Storage
	kiwiCompleteTreeSvc kiwiTree.Service
	kiwiPartialTreeSvc  kiwiTree.Service
}

var (

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")

	// ErrShouldRetry indicates a storage entity is locked and the operation should be retried.
	ErrShouldRetry = errors.New("retry the operation")

	// ErrInternal indicates some unexpected internal failure.
	ErrInternal = errors.New("internal failure")

	// ErrCleanKiwis indicates unused kiwis cleanup failure upon a subscription deletion.
	ErrCleanKiwis = errors.New("kiwis cleanup failure, may cause kiwis garbage")

	// ErrInvalidQuery indicates the search query is invalid.
	ErrInvalidQuery = errors.New("invalid search query")
)

func NewService(
	stor storage.Storage,
	kiwiCompleteTreeSvc kiwiTree.Service,
	kiwiPartialTreeSvc kiwiTree.Service,
) Service {
	return service{
		stor:                stor,
		kiwiCompleteTreeSvc: kiwiCompleteTreeSvc,
		kiwiPartialTreeSvc:  kiwiPartialTreeSvc,
	}
}

func (svc service) Create(ctx context.Context, acc string, sd subscription.Data) (id string, err error) {
	err = sd.Validate()
	if err == nil {
		err = svc.createCondition(ctx, sd.Condition)
		if err == nil {
			id, err = svc.stor.Create(ctx, acc, sd)
		}
	}
	err = translateError(err)
	return
}

func (svc service) createCondition(ctx context.Context, cond condition.Condition) (err error) {
	switch c := cond.(type) {
	case condition.GroupCondition:
		for _, childCond := range c.GetGroup() {
			err = svc.createCondition(ctx, childCond)
			if err != nil {
				break
			}
		}
	case condition.KiwiTreeCondition:
		kiwiTreeSvc := svc.selectKiwiTreeService(c)
		err = kiwiTreeSvc.Create(ctx, c.GetKey(), c.GetPattern())
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", subscription.ErrInvalidSubscriptionCondition, reflect.TypeOf(cond))
	}
	return
}

func (svc service) selectKiwiTreeService(ktc condition.KiwiTreeCondition) (kiwiTreeSvc kiwiTree.Service) {
	if ktc.IsPartial() {
		kiwiTreeSvc = svc.kiwiPartialTreeSvc
	} else {
		kiwiTreeSvc = svc.kiwiCompleteTreeSvc
	}
	return
}

func (svc service) Read(ctx context.Context, id, acc string) (sd subscription.Data, err error) {
	sd, err = svc.stor.Read(ctx, id, acc)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) UpdateMetadata(ctx context.Context, id, acc string, md subscription.Metadata) (err error) {
	err = svc.stor.UpdateMetadata(ctx, id, acc, md)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Delete(ctx context.Context, id, acc string) (err error) {
	var sd subscription.Data
	sd, err = svc.stor.Delete(ctx, id, acc)
	if err == nil {
		err = svc.clearUnusedCondition(ctx, sd.Condition)
		if err != nil {
			err = fmt.Errorf("%w: %s, subscription id: %s", ErrCleanKiwis, err, id)
		}
	}
	err = translateError(err)
	return
}

func (svc service) clearUnusedCondition(ctx context.Context, cond condition.Condition) (err error) {
	switch c := cond.(type) {
	case condition.GroupCondition:
		for _, childCond := range c.GetGroup() {
			err = svc.clearUnusedCondition(ctx, childCond)
			if err != nil {
				break
			}
		}
	case condition.KiwiTreeCondition:
		err = svc.clearUnusedKiwiTreeCondition(ctx, c)
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", subscription.ErrInvalidSubscriptionCondition, reflect.TypeOf(cond))
	}
	return
}

func (svc service) clearUnusedKiwiTreeCondition(ctx context.Context, ktc condition.KiwiTreeCondition) (err error) {
	k := ktc.GetKey()
	p := ktc.GetPattern()
	q := storage.KiwiQuery{
		Key:     k,
		Pattern: p,
	}
	kiwiTreeSvc := svc.selectKiwiTreeService(ktc)
	err = kiwiTreeSvc.LockCreate(ctx, k, p)
	if err == nil {
		defer func() {
			_ = kiwiTreeSvc.UnlockCreate(ctx, k, p)
		}()
		// find any subscription that is also using this kiwi condition
		var matches []*subscription.ConditionMatch
		consumeFunc := func(match *subscription.ConditionMatch) (err error) {
			matches = append(matches, match)
			return
		}
		err = svc.stor.SearchByKiwi(ctx, q, consumeFunc)
		if err == nil {
			if len(matches) == 0 {
				// no other subscriptions found, let's delete the kiwi condition from the tree
				err = kiwiTreeSvc.Delete(ctx, k, p)
			}
		}
	}
	return
}

func (svc service) SearchByAccount(ctx context.Context, q subscription.QueryByAccount, cursor string) (ids []string, err error) {
	ids, err = svc.stor.SearchByAccount(ctx, q, cursor)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) SearchByCondition(ctx context.Context, cond condition.Condition, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	switch condT := cond.(type) {
	case condition.KiwiCondition:
		kiwiQuery := storage.KiwiQuery{
			Key:     condT.GetKey(),
			Pattern: condT.GetPattern(),
			Partial: condT.IsPartial(),
		}
		err = svc.stor.SearchByKiwi(ctx, kiwiQuery, consumeFunc)
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", ErrInvalidQuery, reflect.TypeOf(condT))
	}
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
		case errors.Is(srcErr, storage.ErrNotFound):
			dstErr = fmt.Errorf("%w: %s", ErrNotFound, srcErr)
		case errors.Is(srcErr, storage.ErrInternal):
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		case errors.Is(srcErr, kiwiTree.ErrShouldRetry):
			dstErr = fmt.Errorf("%w: %s", ErrShouldRetry, srcErr)
		case errors.Is(srcErr, kiwiTree.ErrInternal):
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		case errors.Is(srcErr, subscription.ErrInvalidSubscriptionCondition):
			dstErr = srcErr
		case errors.Is(srcErr, ErrNotFound):
			dstErr = srcErr
		case errors.Is(srcErr, ErrInternal):
			dstErr = srcErr
		case errors.Is(srcErr, ErrShouldRetry):
			dstErr = srcErr
		case errors.Is(srcErr, ErrCleanKiwis):
			dstErr = srcErr
		case errors.Is(srcErr, ErrInvalidQuery):
			dstErr = srcErr
		default:
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		}
	}
	return
}
