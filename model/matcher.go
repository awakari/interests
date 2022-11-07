package model

type (

	// Matcher represents a key-pattern matching data.
	Matcher struct {
		MatcherData

		// If true, then allowed match any lexeme in a tokenized metadata value. Otherwise, entire value should match.
		Partial bool
	}
)
