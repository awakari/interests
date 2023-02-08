package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/service/kiwi-tree"
	"github.com/awakari/subscriptions/storage"
	"reflect"
)

// Service is a model.Subscription CRUDL service.
type Service interface {

	// Create a new model.Subscription with the specified model.Data.
	// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
	// Returns model.ErrInvalidSubscriptionRoute if the specified CreateRequest is invalid.
	Create(ctx context.Context, sd subscription.Data) (id string, err error)

	// Read the specified model.Subscription.
	// Returns ErrNotFound if Subscription is missing in the underlying storage.
	Read(ctx context.Context, id string) (sd subscription.Data, err error)

	// Delete a model.Subscription and all associated conditions those not in use by any other model.Subscription.
	// Returns ErrNotFound if model.Subscription with the specified name is missing in the underlying storage.
	Delete(ctx context.Context, id string) (err error)

	// SearchByCondition returns subscription.ConditionMatch page where:<br/>
	// * subscription id is greater than the specified cursor<br/>
	// * contains a condition specified by the query.
	SearchByCondition(ctx context.Context, q condition.Query, cursor string) (page []subscription.ConditionMatch, err error)

	// SearchByMetadata returns all subscriptions those have the metadata matching the query (same keys and values).
	SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []subscription.Subscription, err error)
}

type service struct {
	stor                storage.Storage
	kiwiCompleteTreeSvc kiwiTree.Service
	kiwiPartialTreeSvc  kiwiTree.Service
}

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

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

func (svc service) Create(ctx context.Context, sd subscription.Data) (id string, err error) {
	err = sd.Route.Validate()
	if err == nil {
		err = svc.createCondition(ctx, sd.Route.Condition)
		if err == nil {
			id, err = svc.stor.Create(ctx, sd)
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
		err = fmt.Errorf("%w: unsupported condition type: %s", subscription.ErrInvalidSubscriptionRoute, reflect.TypeOf(cond))
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

func (svc service) Read(ctx context.Context, id string) (sd subscription.Data, err error) {
	sd, err = svc.stor.Read(ctx, id)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Delete(ctx context.Context, id string) (err error) {
	var sd subscription.Data
	sd, err = svc.stor.Delete(ctx, id)
	if err == nil {
		err = svc.clearUnusedCondition(ctx, sd.Route.Condition)
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
		err = fmt.Errorf("%w: unsupported condition type: %s", subscription.ErrInvalidSubscriptionRoute, reflect.TypeOf(cond))
	}
	return
}

func (svc service) clearUnusedKiwiTreeCondition(ctx context.Context, ktc condition.KiwiTreeCondition) (err error) {
	k := ktc.GetKey()
	p := ktc.GetPattern()
	q := storage.KiwiQuery{
		Limit:   1,
		Key:     k,
		Pattern: p,
	}
	var subs []subscription.ConditionMatch
	kiwiTreeSvc := svc.selectKiwiTreeService(ktc)
	err = kiwiTreeSvc.LockCreate(ctx, k, p)
	if err == nil {
		defer func() {
			_ = kiwiTreeSvc.UnlockCreate(ctx, k, p)
		}()
		// find any subscription that is also using this kiwi condition
		subs, err = svc.stor.SearchByKiwi(ctx, q, "")
		if err == nil {
			if len(subs) == 0 {
				// no other subscriptions found, let's delete the kiwi condition from the tree
				err = kiwiTreeSvc.Delete(ctx, k, p)
			}
		}
	}
	return
}

func (svc service) SearchByCondition(ctx context.Context, q condition.Query, cursor string) (page []subscription.ConditionMatch, err error) {
	switch c := q.Condition.(type) {
	case condition.KiwiCondition:
		kiwiQuery := storage.KiwiQuery{
			Limit:   q.Limit,
			Key:     c.GetKey(),
			Pattern: c.GetPattern(),
			Partial: c.IsPartial(),
		}
		page, err = svc.stor.SearchByKiwi(ctx, kiwiQuery, cursor)
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", ErrInvalidQuery, reflect.TypeOf(c))
	}
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) SearchByMetadata(ctx context.Context, q model.MetadataQuery, cursor string) (page []subscription.Subscription, err error) {
	page, err = svc.stor.SearchByMetadata(ctx, q, cursor)
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
		case errors.Is(srcErr, kiwiTree.ErrShouldRetry):
			dstErr = fmt.Errorf("%w: %s", ErrShouldRetry, srcErr)
		case errors.Is(srcErr, kiwiTree.ErrInternal):
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		case errors.Is(srcErr, subscription.ErrInvalidSubscriptionRoute):
			dstErr = srcErr
		case errors.Is(srcErr, ErrNotFound):
			dstErr = srcErr
		case errors.Is(srcErr, ErrInternal):
			dstErr = srcErr
		case errors.Is(srcErr, ErrConflict):
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
