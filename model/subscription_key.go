package model

import (
	"errors"
	"fmt"
)

type (

	// SubscriptionKey is a key part of the Subscription that uniquely defines it.
	SubscriptionKey struct {

		// Name represents a unique subscription name.
		Name string

		// Version represents a Subscription entry version for the optimistic lock purpose.
		Version uint64
	}
)

var (

	// ErrInvalidSubscriptionName indicates the SubscriptionKey.Name is invalid
	ErrInvalidSubscriptionName = errors.New("invalid subscription name")
)

func (sk SubscriptionKey) Validate() (err error) {
	if sk.Name == "" {
		err = fmt.Errorf("%w: %s", ErrInvalidSubscriptionName, "should not be empty")
	}
	return
}
