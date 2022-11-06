package model

import (
	"hash/fnv"
)

type (

	// MatcherData is the part of Matcher stored in the external matchers service.
	MatcherData struct {

		// Key represents the Metadata Key
		Key string

		// Pattern represents a Metadata value matching Pattern
		Pattern Pattern
	}
)

func (md MatcherData) String() string {
	return md.Key + ": " + md.Pattern.String()
}

func (md MatcherData) HashCode() uint64 {
	h := fnv.New64()
	h.Write([]byte(md.Key))
	return h.Sum64() ^ md.Pattern.HashCode()
}
