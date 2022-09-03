package storage

type (

	// Matcher represents a key-pattern matching data.
	Matcher struct {

		// Metadata Key
		Key string

		// Metadata value matching pattern external id
		PatternCode []byte

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
