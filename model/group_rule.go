package model

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
)

type (
	GroupRule interface {
		Rule
		GetLogic() (logic GroupLogic)
		GetGroup() (group []Rule)
		Validate() error
	}

	groupRule struct {
		Rule  Rule
		Logic GroupLogic
		Group []Rule
	}
)

var (

	// ErrInvalidGroupRule indicates the GroupRule is invalid
	ErrInvalidGroupRule = errors.New("invalid group rule")
)

func NewGroupRule(r Rule, logic GroupLogic, group []Rule) GroupRule {
	return groupRule{
		Rule:  r,
		Logic: logic,
		Group: group,
	}
}

func (gr groupRule) IsNot() bool {
	return gr.Rule.IsNot()
}

func (gr groupRule) Equal(another Rule) (equal bool) {
	equal = gr.Rule.Equal(another)
	if equal {
		var anotherGr GroupRule
		anotherGr, equal = another.(GroupRule)
		ruleEqFunc := func(r1, r2 Rule) bool {
			return r1.Equal(r2)
		}
		equal = gr.Logic == anotherGr.GetLogic() && slices.EqualFunc(gr.Group, anotherGr.GetGroup(), ruleEqFunc)
	}
	return
}

func (gr groupRule) GetLogic() (logic GroupLogic) {
	return gr.Logic
}

func (gr groupRule) GetGroup() (group []Rule) {
	return gr.Group
}

func (gr groupRule) Validate() error {
	if len(gr.Group) < 2 {
		return fmt.Errorf("%w: empty or contains single child rule only", ErrInvalidGroupRule)
	}
	for i, r1 := range gr.Group {
		if i < len(gr.Group)-1 {
			for _, r2 := range gr.Group[i+1:] {
				if r1.Equal(r2) {
					return fmt.Errorf("%w: duplicate rule in the group: %v", ErrInvalidGroupRule, r1)
				}
			}
		}
	}
	return nil
}
