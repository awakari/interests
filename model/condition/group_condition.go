package condition

import (
	"slices"
)

type (
	GroupCondition interface {
		Condition
		GetLogic() (logic GroupLogic)
		GetGroup() (group []Condition)
	}

	groupCondition struct {
		Condition Condition
		Logic     GroupLogic
		Group     []Condition
	}
)

func NewGroupCondition(c Condition, logic GroupLogic, group []Condition) GroupCondition {
	return groupCondition{
		Condition: c,
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
