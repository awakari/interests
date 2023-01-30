package model

import (
	"errors"
	"fmt"
)

type (

	// Subscription represents the subscription entry.
	Subscription struct {

		// Name represents the unique Subscription name.
		Name string

		// Description represents an optional human readable Subscription description.
		Description string

		// Routes represents a list of routes associated with the Subscription.
		Routes []string

		// Condition represents the certain criteria to select the Subscription.
		Condition Condition
	}
)

var (

	// ErrInvalidSubscription indicates the Subscription is invalid
	ErrInvalidSubscription = errors.New("invalid subscription")
)

func (sub Subscription) Validate() (err error) {
	if len(sub.Name) == 0 {
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "empty name")
	}
	if len(sub.Routes) == 0 {
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "empty routes")
	}
	if sub.Condition.IsNot() {
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "root condition negation")
	}
	switch c := sub.Condition.(type) {
	case GroupCondition:
		err = validateRootGroupCondition(c)
	case KiwiCondition:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "root condition is not a group neither metadata pattern condition")
	}
	return
}

func validateRootGroupCondition(gc GroupCondition) (err error) {
	err = gc.Validate()
	if err == nil {
		group := gc.GetGroup()
		if len(group) > 1 {
			includesFound := false
			for _, c := range group {
				if !c.IsNot() {
					includesFound = true
				}
				break
			}
			if !includesFound {
				err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "there should be at least 1 includes group in the root condition group")
			}
		} else {
			err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "there should be 1 or 2 child conditions in the root condition group")
		}
	}
	return
}
