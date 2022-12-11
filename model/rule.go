package model

type (
	Rule interface {
		IsNot() bool

		Equal(another Rule) (equal bool)
	}

	rule struct {
		Not bool
	}
)

func NewRule(not bool) Rule {
	return rule{
		Not: not,
	}
}

func (r rule) IsNot() bool {
	return r.Not
}

func (r rule) Equal(another Rule) bool {
	return r.Not == another.IsNot()
}
