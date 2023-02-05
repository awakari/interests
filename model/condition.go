package model

type Condition interface {
	IsNot() bool
	Equal(another Condition) (equal bool)
}

type condition struct {
	Not bool
}

func NewCondition(not bool) Condition {
	return condition{
		Not: not,
	}
}

func (c condition) IsNot() bool {
	return c.Not
}

func (r condition) Equal(another Condition) bool {
	return r.Not == another.IsNot()
}
