package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/service/kiwi-tree"
	"github.com/awakari/subscriptions/storage"
	"reflect"
)

// Service is a model.Subscription CRUDL service.
type Service interface {

	// Create an empty model.Subscription with the specified name and description.
	// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
	// Returns model.ErrInvalidSubscription if the specified CreateRequest is invalid.
	Create(ctx context.Context, sub model.Subscription) (err error)

	// Read the specified model.Subscription.
	// Returns ErrNotFound if Subscription is missing in the underlying storage.
	Read(ctx context.Context, name string) (sub model.Subscription, err error)

	// Delete a model.Subscription and all associated conditions those not in use by any other model.Subscription.
	// Returns ErrNotFound if model.Subscription with the specified name is missing in the underlying storage.
	Delete(ctx context.Context, name string) (err error)

	// ListNames returns all subscription names starting from the specified cursor.
	ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

	// SearchByCondition returns model.Subscription page where:<br/>
	// * name is greater than the specified cursor<br/>
	// * contains a model.Condition specified by the query.
	SearchByCondition(ctx context.Context, q model.ConditionQuery, cursor string) (page []model.Subscription, err error)
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

func (svc service) Create(ctx context.Context, sub model.Subscription) (err error) {
	err = sub.Validate()
	if err == nil {
		err = svc.createCondition(ctx, sub.Condition)
		if err == nil {
			err = svc.stor.Create(ctx, sub)
		}
	}
	err = translateError(err)
	return
}

func (svc service) createCondition(ctx context.Context, cond model.Condition) (err error) {
	switch c := cond.(type) {
	case model.GroupCondition:
		for _, childCond := range c.GetGroup() {
			err = svc.createCondition(ctx, childCond)
			if err != nil {
				break
			}
		}
	case model.KiwiTreeCondition:
		kiwiTreeSvc := svc.selectKiwiTreeService(c)
		err = kiwiTreeSvc.Create(ctx, c.GetKey(), c.GetPattern())
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", model.ErrInvalidSubscription, reflect.TypeOf(cond))
	}
	return
}

func (svc service) selectKiwiTreeService(ktc model.KiwiTreeCondition) (kiwiTreeSvc kiwiTree.Service) {
	if ktc.IsPartial() {
		kiwiTreeSvc = svc.kiwiPartialTreeSvc
	} else {
		kiwiTreeSvc = svc.kiwiCompleteTreeSvc
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
		err = svc.clearUnusedCondition(ctx, sub.Condition)
		if err != nil {
			err = fmt.Errorf("%w: %s, subscription: %v", ErrCleanKiwis, err, sub)
		}
	}
	err = translateError(err)
	return
}

func (svc service) clearUnusedCondition(ctx context.Context, cond model.Condition) (err error) {
	switch c := cond.(type) {
	case model.GroupCondition:
		for _, childCond := range c.GetGroup() {
			err = svc.clearUnusedCondition(ctx, childCond)
			if err != nil {
				break
			}
		}
	case model.KiwiTreeCondition:
		err = svc.clearUnusedKiwiTreeCondition(ctx, c)
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", model.ErrInvalidSubscription, reflect.TypeOf(cond))
	}
	return
}

func (svc service) clearUnusedKiwiTreeCondition(ctx context.Context, ktc model.KiwiTreeCondition) (err error) {
	k := ktc.GetKey()
	p := ktc.GetPattern()
	q := storage.KiwiQuery{
		Limit:   1,
		Key:     k,
		Pattern: p,
	}
	var subs []model.Subscription
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

func (svc service) ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error) {
	page, err = svc.stor.ListNames(ctx, limit, cursor)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) SearchByCondition(ctx context.Context, q model.ConditionQuery, cursor string) (page []model.Subscription, err error) {
	switch c := q.Condition.(type) {
	case model.KiwiCondition:
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
