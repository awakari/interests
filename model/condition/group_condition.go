package condition

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

func NewGroupCondition(c Condition, logic GroupLogic, group []Condition) GroupCondition {
	return groupCondition{
		Condition: c,
		Logic:     logic,
		Group:     group,
	}
}

func (gc groupCondition) GetId() string {
	return gc.Condition.GetId()
}

func (gc groupCondition) IsNot() bool {
	return gc.Condition.IsNot()
}

func (gc groupCondition) Equal(another Condition) (equal bool) {
	equal = gc.Condition.Equal(another)
	if equal {
		var anotherGc GroupCondition
		anotherGc, equal = another.(GroupCondition)
		condEqFunc := func(c1, c2 Condition) bool {
			return c1.Equal(c2)
		}
		equal = gc.Logic == anotherGc.GetLogic() && slices.EqualFunc(gc.Group, anotherGc.GetGroup(), condEqFunc)
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
	containsNegation := false
	allNegations := true
	for i, c1 := range gc.Group {
		if c1.IsNot() {
			if !containsNegation {
				containsNegation = true
			}
		} else if allNegations {
			allNegations = false
		}
		if i < len(gc.Group)-1 {
			for _, c2 := range gc.Group[i+1:] {
				if c1.Equal(c2) {
					return fmt.Errorf("%w: duplicate condition in the group: %v", ErrInvalidGroupCondition, c1)
				}
			}
		}
	}
	if allNegations {
		return fmt.Errorf("%w: contains negation conditions only", ErrInvalidGroupCondition)
	}
	if containsNegation && (gc.Logic == GroupLogicOr || gc.Logic == GroupLogicXor) {
		return fmt.Errorf("%w: contains a negation with group logic %s", ErrInvalidGroupCondition, gc.Logic)
	}
	return nil
}
