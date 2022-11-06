package model

type (

	// Matcher represents a key-pattern matching data.
	Matcher struct {
		MatcherData

		// If true, then allowed match any lexeme in a tokenized metadata value. Otherwise, entire value should match.
		Partial bool
	}
)

func (m Matcher) HashCode() uint64 {
	var partial uint64 = 0
	if m.Partial {
		partial = 1
	}
	return m.MatcherData.HashCode() ^ partial
}
