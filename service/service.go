package service

import (
	"context"
	"errors"
	"fmt"
	conditions_text "github.com/awakari/subscriptions/api/grpc/conditions-text"
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
	Create(ctx context.Context, groupId, userId string, d subscription.Data) (id string, err error)

	// Read the specified subscription.Subscription.
	// Returns ErrNotFound if subscription.Subscription is missing in the underlying storage.
	Read(ctx context.Context, id, groupId, userId string) (d subscription.Data, err error)

	// Update the mutable part of the subscription.Data
	Update(ctx context.Context, id, groupId, userId string, d subscription.Data) (err error)

	// Delete a subscription.Subscription and all associated conditions those not in use by any other subscription.
	// Returns ErrNotFound if a subscription.Subscription with the specified id is missing in the underlying storage.
	Delete(ctx context.Context, id, groupId, userId string) (err error)

	// SearchOwn returns all subscription ids those have the account matching the query.
	SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error)

	// SearchByCondition sends the matches to the specified consumer func those contain a condition specified by the query.
	SearchByCondition(ctx context.Context, condId string, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error)
}

type service struct {
	stor        storage.Storage
	condTextSvc conditions_text.Service
}

var (

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")

	// ErrShouldRetry indicates a storage entity is locked and the operation should be retried.
	ErrShouldRetry = errors.New("retry the operation")

	// ErrInternal indicates some unexpected internal failure.
	ErrInternal = errors.New("internal failure")

	// ErrCleanConditions indicates unused conditions cleanup failure upon a subscription deletion.
	ErrCleanConditions = errors.New("conditions cleanup failure, may cause a garbage in the conditions storage")
)

func NewService(
	stor storage.Storage,
	condTextSvc conditions_text.Service,
) Service {
	return service{
		stor:        stor,
		condTextSvc: condTextSvc,
	}
}

func (svc service) Create(ctx context.Context, groupId, userId string, sd subscription.Data) (id string, err error) {
	err = sd.Validate()
	if err == nil {
		err = svc.createCondition(ctx, sd.Condition)
		if err == nil {
			id, err = svc.stor.Create(ctx, groupId, userId, sd)
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
	case condition.TextCondition:
		var condId, condTerm string
		condId, condTerm, err = svc.condTextSvc.Create(ctx, c.GetKey(), c.GetTerm(), c.IsExact())
		if err == nil {
			c.SetId(condId)
			c.SetTerm(condTerm)
		}
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", subscription.ErrInvalidSubscriptionCondition, reflect.TypeOf(cond))
	}
	return
}

func (svc service) Read(ctx context.Context, id, groupId, userId string) (sd subscription.Data, err error) {
	sd, err = svc.stor.Read(ctx, id, groupId, userId)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Update(ctx context.Context, id, groupId, userId string, d subscription.Data) (err error) {
	err = svc.stor.Update(ctx, id, groupId, userId, d)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) Delete(ctx context.Context, id, groupId, userId string) (err error) {
	var sd subscription.Data
	sd, err = svc.stor.Delete(ctx, id, groupId, userId)
	if err == nil {
		err = svc.clearUnusedCondition(ctx, sd.Condition)
		if err != nil {
			err = fmt.Errorf("%w: %s, subscription id: %s", ErrCleanConditions, err, id)
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
	case condition.TextCondition:
		err = svc.clearUnusedTextCondition(ctx, c.GetId())
	default:
		err = fmt.Errorf("%w: unsupported condition type: %s", subscription.ErrInvalidSubscriptionCondition, reflect.TypeOf(cond))
	}
	return
}

func (svc service) clearUnusedTextCondition(ctx context.Context, condId string) (err error) {
	err = svc.condTextSvc.LockCreate(ctx, condId)
	if err == nil {
		defer svc.condTextSvc.UnlockCreate(ctx, condId)
		// find any subscription that is also using this condition
		var matches []*subscription.ConditionMatch
		consumeFunc := func(match *subscription.ConditionMatch) (err error) {
			matches = append(matches, match)
			return
		}
		err = svc.stor.SearchByCondition(ctx, condId, consumeFunc)
		if err == nil {
			if len(matches) == 0 {
				// no other subscriptions found, let's delete the condition from the tree
				err = svc.condTextSvc.Delete(ctx, condId)
			}
		}
	}
	return
}

func (svc service) SearchOwn(ctx context.Context, q subscription.QueryOwn, cursor string) (ids []string, err error) {
	ids, err = svc.stor.SearchOwn(ctx, q, cursor)
	if err != nil {
		err = translateError(err)
	}
	return
}

func (svc service) SearchByCondition(ctx context.Context, condId string, consumeFunc util.ConsumeFunc[*subscription.ConditionMatch]) (err error) {
	err = svc.stor.SearchByCondition(ctx, condId, consumeFunc)
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
		case errors.Is(srcErr, conditions_text.ErrConflict):
			dstErr = fmt.Errorf("%w: %s", ErrShouldRetry, srcErr)
		case errors.Is(srcErr, subscription.ErrInvalidSubscriptionCondition):
			dstErr = srcErr
		case errors.Is(srcErr, ErrNotFound):
			dstErr = srcErr
		case errors.Is(srcErr, ErrInternal):
			dstErr = srcErr
		case errors.Is(srcErr, ErrShouldRetry):
			dstErr = srcErr
		case errors.Is(srcErr, ErrCleanConditions):
			dstErr = srcErr
		default:
			dstErr = fmt.Errorf("%w: %s", ErrInternal, srcErr)
		}
	}
	return
}
