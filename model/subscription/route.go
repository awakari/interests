package subscription

import (
	"errors"
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
)

// Route represents the Subscription routing data that includes a routing Condition and destination routes.
type Route struct {

	// Destinations represents a list of destination routes associated with the Subscription.
	// A message is delivered to all the Destinations when Condition matches the message.
	Destinations []string

	// Condition represents the certain criteria to select the Subscription for the further routing.
	Condition condition.Condition
}

// ErrInvalidSubscriptionRoute indicates the Subscription is invalid
var ErrInvalidSubscriptionRoute = errors.New("invalid subscription route")

func (sub Route) Validate() (err error) {
	if len(sub.Destinations) == 0 {
		return fmt.Errorf("%w: %s", ErrInvalidSubscriptionRoute, "empty destinations")
	}
	if sub.Condition.IsNot() {
		return fmt.Errorf("%w: %s", ErrInvalidSubscriptionRoute, "root condition negation")
	}
	switch c := sub.Condition.(type) {
	case condition.GroupCondition:
		err = validateRootGroupCondition(c)
	case condition.KiwiCondition:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSubscriptionRoute, "root condition is not a group neither metadata pattern condition")
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
				err = fmt.Errorf("%w: %s", ErrInvalidSubscriptionRoute, "there should be at least 1 includes group in the root condition group")
			}
		} else {
			err = fmt.Errorf("%w: %s", ErrInvalidSubscriptionRoute, "there should be 1 or 2 child conditions in the root condition group")
		}
	}
	return
}
