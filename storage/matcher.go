package storage

import "subscriptions/patterns"

type (
	// Matcher represents a key-pattern matching data.
	Matcher struct {

		// Metadata Key
		Key string

		// Metadata value matching pattern external id
		PatternCode patterns.Code

		// If true, then allowed match any lexeme in a tokenized metadata value. Otherwise, entire value should match.
		Partial bool
	}
)
