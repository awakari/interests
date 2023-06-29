package subscription

import "github.com/awakari/subscriptions/model/condition"

// ConditionMatch represents a subscription that contains a condition with the matching id.
type ConditionMatch struct {
	SubscriptionId string

	// Condition represents the root Subscription condition.
	Condition condition.Condition
}
