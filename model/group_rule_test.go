package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupRule_IsNot(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	gr1 := NewGroupRule(r1, GroupAnd, []Rule{})
	gr2 := NewGroupRule(r2, GroupAnd, []Rule{})
	assert.False(t, gr1.IsNot())
	assert.True(t, gr2.IsNot())
}

func TestGroupRule_GetLogic(t *testing.T) {
	r1 := NewRule(false)
	gr1 := NewGroupRule(r1, GroupAnd, []Rule{})
	gr2 := NewGroupRule(r1, GroupOr, []Rule{})
	assert.Equal(t, GroupAnd, gr1.GetLogic())
	assert.Equal(t, GroupOr, gr2.GetLogic())
}

func TestGroupRule_GetGroup(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	group := []Rule{
		r1,
		r2,
	}
	gr1 := NewGroupRule(r1, GroupAnd, group)
	assert.ElementsMatch(t, group, gr1.GetGroup())
}

func TestGroupRule_Equal(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	gr1 := NewGroupRule(r1, GroupAnd, []Rule{r1, r2})
	cases := map[string]struct {
		in  GroupRule
		out bool
	}{
		"equals": {
			in:  NewGroupRule(r1, GroupAnd, []Rule{r1, r2}),
			out: true,
		},
		"different base rule": {
			in: NewGroupRule(r2, GroupAnd, []Rule{r1, r2}),
		},
		"different logic": {
			in: NewGroupRule(r1, GroupXor, []Rule{r1, r2}),
		},
		"different child rule group": {
			in: NewGroupRule(r1, GroupAnd, []Rule{r2}),
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := gr1.Equal(c.in)
			assert.Equal(t, c.out, equal)
		})
	}
}

func TestGroupRule_Validate(t *testing.T) {
	r1 := NewRule(false)
	r2 := NewRule(true)
	cases := map[string]struct {
		in  GroupRule
		err error
	}{
		"ok": {
			in: NewGroupRule(r1, GroupAnd, []Rule{r1, r2}),
		},
		"empty": {
			in:  NewGroupRule(r1, GroupAnd, []Rule{}),
			err: ErrInvalidGroupRule,
		},
		"single child rule": {
			in:  NewGroupRule(r1, GroupAnd, []Rule{r2}),
			err: ErrInvalidGroupRule,
		},
		"duplicate child rule": {
			in:  NewGroupRule(r1, GroupAnd, []Rule{r2, r2}),
			err: ErrInvalidGroupRule,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := c.in.Validate()
			assert.ErrorIs(t, err, c.err)
		})
	}
}
