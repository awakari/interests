package condition

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyCondition_IsNot(t *testing.T) {
	r1 := NewCondition("", false)
	r2 := NewCondition("", true)
	mdr1 := NewKeyCondition(r1, "")
	mdr2 := NewKeyCondition(r2, "")
	assert.False(t, mdr1.IsNot())
	assert.True(t, mdr2.IsNot())
}

func TestKeyCondition_GetKey(t *testing.T) {
	r1 := NewCondition("", false)
	mdr1 := NewKeyCondition(r1, "metadata key")
	assert.Equal(t, "metadata key", mdr1.GetKey())
}

func TestKeyCondition_Equal(t *testing.T) {
	r1 := NewCondition("", false)
	r2 := NewCondition("", true)
	mdr1 := NewKeyCondition(r1, "key1")
	cases := map[string]struct {
		in    KeyCondition
		equal bool
	}{
		"equals": {
			in:    mdr1,
			equal: true,
		},
		"different base condition": {
			in: NewKeyCondition(r2, "key1"),
		},
		"different key": {
			in: NewKeyCondition(r1, "key2"),
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := mdr1.Equal(c.in)
			assert.Equal(t, c.equal, equal)
		})
	}
}
