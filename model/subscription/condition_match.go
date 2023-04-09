package subscription

import "github.com/awakari/subscriptions/model/condition"

// ConditionMatch represents a subscription id and a condition id that matched.
type ConditionMatch struct {
	SubscriptionId string

	// ConditionId represents the id of a Condition that matched.
	ConditionId string

	// Condition represents the root Subscription condition.
	Condition condition.Condition
}
