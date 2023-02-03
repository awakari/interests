package model

import "github.com/google/uuid"

type Condition interface {
	GetId() string

	IsNot() bool

	Equal(another Condition) (equal bool)
}

type condition struct {
	Id  string
	Not bool
}

func NewCondition(not bool) Condition {
	return NewConditionWithId(not, uuid.NewString())
}

func NewConditionWithId(not bool, id string) Condition {
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

func (c condition) Equal(another Condition) bool {
	return c.Id == another.GetId()
}
