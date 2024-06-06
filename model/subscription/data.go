package subscription

import (
	"github.com/awakari/subscriptions/model/condition"
	"time"
)

type Data struct {

	// Description is human readable subscription description
	Description string

	// Enabled defines whether the subscription may be used for the matching
	Enabled bool

	// Expires defines a deadline when the subscription is treated as Enabled
	Expires time.Time

	// Created represents the subscription creation time.
	Created time.Time

	// Updated represents the subscription last update time.
	Updated time.Time

	// Read represents the latest observed time when a message matching the subscription has been read
	Read time.Time

	// Condition represents the certain criteria to select the Subscription for the further routing.
	// It's immutable once the Subscription is created.
	Condition condition.Condition
}
