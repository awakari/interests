package model

// KiwiTreeCondition is a marker type representing that the pattern is/to be stored in the kiwi-tree storage.
type KiwiTreeCondition interface {
	KiwiCondition
}

type kiwiTreeCondition struct {
	kiwiCondition
}

func NewKiwiTreeCondition(kc KiwiCondition) KiwiTreeCondition {
	return kiwiTreeCondition{
		kiwiCondition{
			KeyCondition: NewKeyCondition(
				NewConditionWithId(kc.IsNot(), kc.GetId()),
				kc.GetKey(),
			),
			Partial: kc.IsPartial(),
			Pattern: kc.GetPattern(),
		},
	}
}
