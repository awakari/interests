package model

type (
	Condition interface {
		IsNot() bool

		Equal(another Condition) (equal bool)
	}

	condition struct {
		Not bool
	}
)

func NewCondition(not bool) Condition {
	return condition{
		Not: not,
	}
}

func (r condition) IsNot() bool {
	return r.Not
}

func (r condition) Equal(another Condition) bool {
	return r.Not == another.IsNot()
}
