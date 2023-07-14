package subscription

import (
	"github.com/awakari/subscriptions/model/condition"
)

type Data struct {
	Description string
	Enabled     bool

	// Condition represents the certain criteria to select the Subscription for the further routing.
	// It's immutable once the Subscription is created.
	Condition condition.Condition
}
