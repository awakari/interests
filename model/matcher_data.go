package model

type (

	// MatcherData is the part of Matcher stored in the external matchers service.
	MatcherData struct {

		// Key represents the Metadata Key
		Key string

		// Pattern represents a Metadata value matching Pattern
		Pattern Pattern
	}
)
