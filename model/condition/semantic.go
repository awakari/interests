package condition

type SemanticCondition interface {
	LeafCondition
	Query() string
}

type semCond struct {
	cond  Condition
	id    string
	query string
}

func NewSemanticCondition(cond Condition, id string, query string) SemanticCondition {
	return semCond{
		cond:  cond,
		id:    id,
		query: query,
	}
}

func (sc semCond) IsNot() bool {
	return sc.cond.IsNot()
}

func (sc semCond) Equal(another Condition) (equal bool) {
	equal = sc.cond.Equal(another)
	var anotherSc SemanticCondition
	if equal {
		anotherSc, equal = another.(SemanticCondition)
	}
	if equal {
		equal = sc.query == anotherSc.Query()
	}
	return
}

func (sc semCond) GetId() string {
	return sc.id
}

func (sc semCond) Query() string {
	return sc.query
}
