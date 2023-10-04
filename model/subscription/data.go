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

	// Condition represents the certain criteria to select the Subscription for the further routing.
	// It's immutable once the Subscription is created.
	Condition condition.Condition
}
