package model

type (

	// Matcher represents a key-pattern matching data.
	Matcher struct {

		// Key represents the Metadata Key
		Key string

		// Pattern represents a Metadata value matching Pattern
		Pattern Pattern

		// If true, then allowed match any lexeme in a tokenized metadata value. Otherwise, entire value should match.
		Partial bool
	}
)
