package condition

type LeafCondition interface {
	Condition
	GetId() string
}
