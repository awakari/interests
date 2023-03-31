package subscription

import "github.com/awakari/subscriptions/model/condition"

// ConditionMatch represents a subscription id and a condition id that matched.
type ConditionMatch struct {
	Key ConditionMatchKey

	// Account represents the Subscription owner id.
	Account string

	// ConditionId represents the id of a Condition that matched.
	ConditionId string

	// Condition represents the root Subscription condition.
	Condition condition.Condition
}

// ConditionMatchKey represents the combination of hash and sorting keys.
type ConditionMatchKey struct {

	// Id represents the unique Subscription.Id. This is a hash key.
	Id string

	// Priority represents the match selection priority for a search query. This is a sorting key.
	Priority uint32
}
