package condition

type TextCondition interface {
	KeyCondition
	GetTerm() string
}

type textCondition struct {
	KeyCondition KeyCondition
	Term         string
}

func NewTextCondition(kc KeyCondition, term string) TextCondition {
	return textCondition{
		KeyCondition: kc,
		Term:         term,
	}
}

func (tc textCondition) IsNot() bool {
	return tc.KeyCondition.IsNot()
}

func (tc textCondition) Equal(another Condition) (equal bool) {
	equal = tc.KeyCondition.Equal(another)
	if equal {
		var anotherTc TextCondition
		anotherTc, equal = another.(TextCondition)
		if equal {
			equal = tc.Term == anotherTc.GetTerm()
		}
	}
	return
}

func (tc textCondition) GetId() string {
	return tc.KeyCondition.GetId()
}

func (tc textCondition) SetId(id string) {
	tc.KeyCondition.SetId(id)
}

func (tc textCondition) GetKey() string {
	return tc.KeyCondition.GetKey()
}

func (tc textCondition) GetTerm() string {
	return tc.Term
}
