package model

import (
	"hash/fnv"
)

type (

	// PatternCode is a pattern identifier. Generally, not equal to the source pattern string.
	PatternCode []byte

	Pattern struct {
		Code PatternCode
		Src  string
	}
)

func (p Pattern) String() string {
	return p.Src
}

func (p Pattern) HashCode() uint64 {
	h := fnv.New64()
	h.Write(p.Code)
	return h.Sum64()
}
