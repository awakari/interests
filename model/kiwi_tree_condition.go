package model

// KiwiTreeCondition is a marker type representing that the pattern is/to be stored in the kiwi-tree storage.
type KiwiTreeCondition interface {
	KiwiCondition
}

type kiwiTreeCondition struct {
	kiwiCondition
}

func NewKiwiTreeCondition(kc kiwiCondition) KiwiTreeCondition {
	return kiwiTreeCondition{
		kc,
	}
}
