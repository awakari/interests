package model

import "bytes"

type (

	// MatcherGroup represent a set of Matcher linked together with some option.
	MatcherGroup struct {

		// All defines whether "all" matchers in the group should match or "any" is sufficient.
		All bool

		// Matchers represents a set of Matcher in the group.
		Matchers []Matcher
	}
)

func (mg MatcherGroup) Matches(md Metadata, key string, patternCode PatternCode) (matches bool, err error) {
	for _, m := range mg.Matchers {
		if key == m.Key && bytes.Equal(patternCode, m.Pattern.Code) {
			matches = true // matched before (key, patternCode) pair
		} else {
			matches, err = m.Matches(md)
			if err != nil {
				break
			}
		}
		if matches {
			if !mg.All {
				break // any match is enough
			}
		} else if mg.All {
			break // any mismatch is enough
		}
	}
	return
}
