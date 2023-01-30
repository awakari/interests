package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRule_IsNot(t *testing.T) {
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	assert.False(t, r1.IsNot())
	assert.True(t, r2.IsNot())
}

func TestRule_Equal(t *testing.T) {
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	r3 := NewCondition(false)
	assert.False(t, r1.Equal(r2))
	assert.True(t, r1.Equal(r3))
}
