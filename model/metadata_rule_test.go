package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetadataRule_IsNot(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	mdr1 := NewMetadataRule(r1, "")
	mdr2 := NewMetadataRule(r2, "")
	assert.False(t, mdr1.IsNot())
	assert.True(t, mdr2.IsNot())
}

func TestMetadataRule_GetKey(t *testing.T) {
	r1 := NewRule(false)
	mdr1 := NewMetadataRule(r1, "metadata key")
	assert.Equal(t, "metadata key", mdr1.GetKey())
}

func TestMetadataRule_Equal(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	mdr1 := NewMetadataRule(r1, "key1")
	cases := map[string]struct {
		in    MetadataRule
		equal bool
	}{
		"equals": {
			in:    mdr1,
			equal: true,
		},
		"different base rule": {
			in: NewMetadataRule(r2, "key1"),
		},
		"different key": {
			in: NewMetadataRule(r1, "key2"),
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := mdr1.Equal(c.in)
			assert.Equal(t, c.equal, equal)
		})
	}
}
