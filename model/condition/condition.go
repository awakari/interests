package condition

type Condition interface {
	GetId() string
	IsNot() bool
	Equal(another Condition) (equal bool)
}

type condition struct {
	Id  string
	Not bool
}

func NewCondition(id string, not bool) Condition {
	return condition{
		Id:  id,
		Not: not,
	}
}

func (c condition) GetId() string {
	return c.Id
}

func (c condition) IsNot() bool {
	return c.Not
}

func (r condition) Equal(another Condition) bool {
	return r.Not == another.IsNot()
}
