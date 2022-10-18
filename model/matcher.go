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

func (m Matcher) Matches(md Metadata) (matches bool) {
	var input string
	input, matches = md[m.Key]
	if matches {

	}
	return
}
