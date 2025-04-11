package condition

type SemanticCondition interface {
	LeafCondition
	Query() string
	SimilarityMin() float32
}

type semCond struct {
	cond          Condition
	id            string
	query         string
	similarityMin float32
}

func NewSemanticCondition(cond Condition, id string, query string, similarityMin float32) SemanticCondition {
	return semCond{
		cond:          cond,
		id:            id,
		query:         query,
		similarityMin: similarityMin,
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
		equal = sc.query == anotherSc.Query() && sc.similarityMin == anotherSc.SimilarityMin()
	}
	return
}

func (sc semCond) GetId() string {
	return sc.id
}

func (sc semCond) Query() string {
	return sc.query
}

func (sc semCond) SimilarityMin() float32 {
	return sc.similarityMin
}
