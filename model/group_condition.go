package model

import (
	"errors"
	"fmt"
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
	return gc.Condition.Equal(another)
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
