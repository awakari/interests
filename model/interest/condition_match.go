package interest

import (
	"github.com/awakari/interests/model/condition"
	"time"
)

// ConditionMatch represents an interest that contains a condition with the matching id.
type ConditionMatch struct {
	InterestId string

	// Condition represents the root Interest condition.
	Condition condition.Condition
}

type ConditionMatchPage struct {
	ConditionMatches []ConditionMatch
	Expires          time.Time
}
