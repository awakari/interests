package model

import (
	"errors"
	"fmt"
)

// Subscription represents the subscription entry.
type Subscription struct {

	// Id represents the unique Subscription Id.
	Id string

	// Data represents the Subscription payload data.
	Data SubscriptionData
}

// SubscriptionData represents the Subscription payload data.
type SubscriptionData struct {

	// Metadata represents the optional subscription attributes, e.g. human-readable description, user ownership.
	Metadata map[string]string

	// Routes represents a list of routes associated with the Subscription.
	Routes []string

	// Condition represents the certain criteria to select the Subscription.
	Condition Condition
}

// ErrInvalidSubscription indicates the Subscription is invalid
var ErrInvalidSubscription = errors.New("invalid subscription")

func (sub SubscriptionData) Validate() (err error) {
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
