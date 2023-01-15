package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupRule_IsNot(t *testing.T) {
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	gr1 := NewGroupCondition(r1, GroupLogicAnd, []Condition{})
	gr2 := NewGroupCondition(r2, GroupLogicAnd, []Condition{})
	assert.False(t, gr1.IsNot())
	assert.True(t, gr2.IsNot())
}

func TestGroupRule_GetLogic(t *testing.T) {
	r1 := NewCondition(false)
	gr1 := NewGroupCondition(r1, GroupLogicAnd, []Condition{})
	gr2 := NewGroupCondition(r1, GroupLogicOr, []Condition{})
	assert.Equal(t, GroupLogicAnd, int(gr1.GetLogic()))
	assert.Equal(t, GroupLogicOr, int(gr2.GetLogic()))
}

func TestGroupRule_GetGroup(t *testing.T) {
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	group := []Condition{
		r1,
		r2,
	}
	gr1 := NewGroupCondition(r1, GroupLogicAnd, group)
	assert.ElementsMatch(t, group, gr1.GetGroup())
}

func TestGroupRule_Equal(t *testing.T) {
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	gr1 := NewGroupCondition(r1, GroupLogicAnd, []Condition{r1, r2})
	cases := map[string]struct {
		in  GroupCondition
		out bool
	}{
		"equals": {
			in:  NewGroupCondition(r1, GroupLogicAnd, []Condition{r1, r2}),
			out: true,
		},
		"different base condition": {
			in: NewGroupCondition(r2, GroupLogicAnd, []Condition{r1, r2}),
		},
		"different logic": {
			in: NewGroupCondition(r1, GroupLogicXor, []Condition{r1, r2}),
		},
		"different child condition group": {
			in: NewGroupCondition(r1, GroupLogicAnd, []Condition{r2}),
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
	r1 := NewCondition(false)
	r2 := NewCondition(true)
	cases := map[string]struct {
		in  GroupCondition
		err error
	}{
		"ok": {
			in: NewGroupCondition(r1, GroupLogicAnd, []Condition{r1, r2}),
		},
		"empty": {
			in:  NewGroupCondition(r1, GroupLogicAnd, []Condition{}),
			err: ErrInvalidGroupCondition,
		},
		"single child condition": {
			in:  NewGroupCondition(r1, GroupLogicAnd, []Condition{r2}),
			err: ErrInvalidGroupCondition,
		},
		"duplicate child condition": {
			in:  NewGroupCondition(r1, GroupLogicAnd, []Condition{r2, r2}),
			err: ErrInvalidGroupCondition,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := c.in.Validate()
			assert.ErrorIs(t, err, c.err)
		})
	}
}
