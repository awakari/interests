package model

// KiwiQuery is the model.Subscription search query by a certain key/pattern ("kiwi") condition pair.
type KiwiQuery struct {

	// Limit defines a results page size limit.
	Limit uint32

	// Key represents the subscription matching criteria: where any model.KiwiCondition has the equal Key.
	Key string

	// Pattern represents the subscription matching criteria: where any model.KiwiCondition has the equal Pattern.
	Pattern string

	// Partial represents the subscription matching criteria: where any model.KiwiCondition has the equal Partial value.
	Partial bool
}
