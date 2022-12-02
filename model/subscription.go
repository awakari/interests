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

		// Includes represents a MatcherGroup to include the Subscription to query results.
		Includes MatcherGroup

		// Excludes represents a MatcherGroup to exclude the Subscription from the query results.
		Excludes MatcherGroup
	}
)

var (

	// ErrInvalidSubscription indicates the Subscription is invalid
	ErrInvalidSubscription = errors.New("invalid subscription")
)

func (sub Subscription) Validate() (err error) {
	if len(sub.Name) == 0 {
		err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "empty name")
	} else if len(sub.Routes) == 0 {
		err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "empty routes")
	} else if len(sub.Includes.Matchers) == 0 && len(sub.Excludes.Matchers) == 0 {
		err = fmt.Errorf("%w: %s", ErrInvalidSubscription, "both includes and excludes matcher groups are empty")
	} else {
		err = sub.Includes.Validate()
		if err != nil {
			err = fmt.Errorf("%w: includes: %s", ErrInvalidSubscription, err)
		} else {
			err = sub.Excludes.Validate()
			if err != nil {
				err = fmt.Errorf("%w: excludes: %s", ErrInvalidSubscription, err)
			}
		}
	}
	return
}
