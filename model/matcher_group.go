package model

import (
	"errors"
	"fmt"
)

type (

	// MatcherGroup represent a set of Matcher linked together with some option.
	MatcherGroup struct {

		// All defines whether "all" matchers in the group are required match or "any" is sufficient.
		All bool

		// Matchers represents a set of Matcher in the group.
		Matchers []Matcher
	}
)

var (

	// ErrInvalidMatcherGroup indicates the MatcherGroup is invalid
	ErrInvalidMatcherGroup = errors.New("invalid matcher group")
)

func (mg MatcherGroup) Validate() error {
	for i, m1 := range mg.Matchers {
		if i < len(mg.Matchers)-1 {
			for _, m2 := range mg.Matchers[i+1:] {
				if m1.Equal(m2) {
					return fmt.Errorf("%w: duplicate matchers: %v", ErrInvalidMatcherGroup, m1)
				}
			}
		}
	}
	return nil
}
