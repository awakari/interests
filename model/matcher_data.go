package model

type (

	// MatcherData is the part of Matcher stored in the external matchers service.
	MatcherData struct {

		// Key represents the metadata Key
		Key string

		// Pattern represents a metadata value matching Pattern
		Pattern Pattern
	}
)

func (md MatcherData) Equal(another MatcherData) bool {
	return md.Key == another.Key && md.Pattern.Equal(another.Pattern)
}
