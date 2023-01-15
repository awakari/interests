package model

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
)

type (
	GroupCondition interface {
		Condition
		GetLogic() (logic GroupLogic)
		GetGroup() (group []Condition)
		Validate() error
	}

	groupCondition struct {
		Condition Condition
		Logic     GroupLogic
		Group     []Condition
	}
)

var (

	// ErrInvalidGroupCondition indicates the GroupCondition is invalid
	ErrInvalidGroupCondition = errors.New("invalid group condition")
)

func NewGroupCondition(r Condition, logic GroupLogic, group []Condition) GroupCondition {
	return groupCondition{
		Condition: r,
		Logic:     logic,
		Group:     group,
	}
}

func (gc groupCondition) IsNot() bool {
	return gc.Condition.IsNot()
}

func (gc groupCondition) Equal(another Condition) (equal bool) {
	equal = gc.Condition.Equal(another)
	if equal {
		var anotherGc GroupCondition
		anotherGc, equal = another.(GroupCondition)
		ruleEqFunc := func(c1, c2 Condition) bool {
			return c1.Equal(c2)
		}
		equal = gc.Logic == anotherGc.GetLogic() && slices.EqualFunc(gc.Group, anotherGc.GetGroup(), ruleEqFunc)
	}
	return
}

func (gc groupCondition) GetLogic() (logic GroupLogic) {
	return gc.Logic
}

func (gc groupCondition) GetGroup() (group []Condition) {
	return gc.Group
}

func (gc groupCondition) Validate() error {
	if len(gc.Group) < 2 {
		return fmt.Errorf("%w: empty or contains single child condition only", ErrInvalidGroupCondition)
	}
	for i, c1 := range gc.Group {
		if i < len(gc.Group)-1 {
			for _, c2 := range gc.Group[i+1:] {
				if c1.Equal(c2) {
					return fmt.Errorf("%w: duplicate condition in the group: %v", ErrInvalidGroupCondition, c1)
				}
			}
		}
	}
	return nil
}
