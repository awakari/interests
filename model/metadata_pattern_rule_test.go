package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetadataPatternRule_IsNot(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	mdpr1 := NewMetadataPatternRule(NewMetadataRule(r1, ""), false, Pattern{})
	mdpr2 := NewMetadataPatternRule(NewMetadataRule(r2, ""), false, Pattern{})
	assert.False(t, mdpr1.IsNot())
	assert.True(t, mdpr2.IsNot())
}

func TestMetadataPatternRule_GetKey(t *testing.T) {
	r1 := NewRule(false)
	mdpr1 := NewMetadataPatternRule(NewMetadataRule(r1, "metadata key"), false, Pattern{})
	assert.Equal(t, "metadata key", mdpr1.GetKey())
}

func TestMetadataPatternRule_GetPattern(t *testing.T) {
	r1 := NewRule(false)
	p1 := Pattern{
		Src: "pattern1",
	}
	mdpr1 := NewMetadataPatternRule(NewMetadataRule(r1, "metadata key"), false, p1)
	assert.Equal(t, p1, mdpr1.GetPattern())
}

func TestMetadataPatternRule_Equal(t *testing.T) {
	r1 := NewRule(false)
	p1 := Pattern{
		Src: "pattern1",
	}
	mdr1 := NewMetadataRule(r1, "key1")
	mdpr1 := NewMetadataPatternRule(mdr1, true, p1)
	cases := map[string]struct {
		in    MetadataPatternRule
		equal bool
	}{
		"equals to itself": {
			in:    mdpr1,
			equal: true,
		},
		"different base rule": {
			in: NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key1"), true, p1),
		},
		"different key": {
			in: NewMetadataPatternRule(NewMetadataRule(r1, "key2"), true, p1),
		},
		"different partial flag": {
			in: NewMetadataPatternRule(mdr1, false, p1),
		},
		"different pattern": {
			in: NewMetadataPatternRule(mdr1, true, Pattern{Src: "pattern2"}),
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := mdpr1.Equal(c.in)
			assert.Equal(t, c.equal, equal)
		})
	}
}
