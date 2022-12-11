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

		// Rule represents the certain criteria to select the Subscription.
		Rule Rule
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
	if sub.Rule.IsNot() {
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "root rule negation")
	}
	switch r := sub.Rule.(type) {
	case GroupRule:
		err = validateRootGroupRule(r)
	case MetadataPatternRule:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "root rule is not a group neither metadata pattern rule")
	}
	return
}

func validateRootGroupRule(gr GroupRule) (err error) {
	err = gr.Validate()
	if err == nil {
		group := gr.GetGroup()
		if len(group) < 3 {
			includesFound := false
			for _, r := range group {
				if !r.IsNot() {
					includesFound = true
				}
				err = validateChildRule(r)
				if err != nil {
					break
				}
			}
			if !includesFound {
				err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "there should be at least 1 includes group in the root rule group")
			}
		} else {
			err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "there should be 1 or 2 child rules in the root rule group")
		}
	}
	return
}

func validateChildRule(r Rule) (err error) {
	switch rr := r.(type) {
	case GroupRule:
		err = rr.Validate()
		if err == nil {
			group := rr.GetGroup()
			for _, childRule := range group {
				_, ok := childRule.(MetadataPatternRule)
				if !ok {
					err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "a child rule group may contain only metadata pattern rules")
					break
				}
			}
		}
	case MetadataPatternRule:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSubscription, "child rule is not a group neither metadata pattern rule")
	}
	return
}
