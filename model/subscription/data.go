package subscription

import (
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
)

type Data struct {

	// Metadata represents a mutable Subscription data.
	Metadata Metadata

	// Condition represents the certain criteria to select the Subscription for the further routing.
	// It's immutable once the Subscription is created.
	Condition condition.Condition
}

// ErrInvalidSubscriptionCondition indicates the Subscription Condition is invalid
var ErrInvalidSubscriptionCondition = errors.New("invalid subscription condition")

func (d Data) Validate() (err error) {
	if d.Condition.IsNot() {
		return fmt.Errorf("%w: %s", ErrInvalidSubscriptionCondition, "root condition negation")
	}
	switch c := d.Condition.(type) {
	case condition.GroupCondition:
		err = validateRootGroupCondition(c)
	case condition.KiwiCondition:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSubscriptionCondition, "root condition is not a group neither metadata pattern condition")
	}
	return
}

func validateRootGroupCondition(gc condition.GroupCondition) (err error) {
	err = gc.Validate()
	if err == nil {
		group := gc.GetGroup()
		if len(group) > 1 {
			includesFound := false
			for _, c := range group {
				if !c.IsNot() {
					includesFound = true
					break
				}
			}
			if !includesFound {
				err = fmt.Errorf("%w: %s", ErrInvalidSubscriptionCondition, "there should be at least 1 includes group in the root condition group")
			}
		} else {
			err = fmt.Errorf("%w: %s", ErrInvalidSubscriptionCondition, "there should be 1 or 2 child conditions in the root condition group")
		}
	}
	return
}
