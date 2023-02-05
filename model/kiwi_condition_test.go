package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetadataPatternRule_IsNot(t *testing.T) {
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	mdpr1 := NewKiwiCondition(NewKeyCondition(r1, "", ""), false, "")
	mdpr2 := NewKiwiCondition(NewKeyCondition(r2, "", ""), false, "")
	assert.False(t, mdpr1.IsNot())
	assert.True(t, mdpr2.IsNot())
}

func TestMetadataPatternRule_GetKey(t *testing.T) {
	r1 := NewCondition(false)
	mdpr1 := NewKiwiCondition(NewKeyCondition(r1, "", "metadata key"), false, "")
	assert.Equal(t, "metadata key", mdpr1.GetKey())
}

func TestMetadataPatternRule_GetPattern(t *testing.T) {
	r1 := NewCondition(false)
	p1 := "pattern1"
	mdpr1 := NewKiwiCondition(NewKeyCondition(r1, "", "metadata key"), false, p1)
	assert.Equal(t, p1, mdpr1.GetPattern())
}

func TestMetadataPatternRule_Equal(t *testing.T) {
	r1 := NewCondition(false)
	p1 := "pattern1"
	mdr1 := NewKeyCondition(r1, "", "key1")
	mdpr1 := NewKiwiCondition(mdr1, true, p1)
	cases := map[string]struct {
		in    KiwiCondition
		equal bool
	}{
		"equals to itself": {
			in:    mdpr1,
			equal: true,
		},
		"different base condition": {
			in: NewKiwiCondition(NewKeyCondition(NewCondition(true), "", "key1"), true, p1),
		},
		"different key": {
			in: NewKiwiCondition(NewKeyCondition(r1, "", "key2"), true, p1),
		},
		"different partial flag": {
			in: NewKiwiCondition(mdr1, false, p1),
		},
		"different pattern": {
			in: NewKiwiCondition(mdr1, true, "pattern2"),
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := mdpr1.Equal(c.in)
			assert.Equal(t, c.equal, equal)
		})
	}
}
