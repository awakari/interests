package storage

type (

	// MatcherGroup represent a set of Matcher linked together with some option.
	MatcherGroup struct {

		// All defines whether "all" matchers in the group should match or "any" is sufficient.
		All bool

		// Matchers represents a set of Matcher in the group.
		Matchers []Matcher
	}
)
