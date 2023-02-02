package model

type (
	KeyCondition interface {
		Condition
		GetKey() string
	}

	keyCondition struct {
		Condition Condition
		Key       string
	}
)

func NewKeyCondition(c Condition, k string) KeyCondition {
	return keyCondition{
		Condition: c,
		Key:       k,
	}
}

func (kc keyCondition) GetId() string {
	return kc.Condition.GetId()
}

func (kc keyCondition) IsNot() bool {
	return kc.Condition.IsNot()
}

func (kc keyCondition) Equal(another Condition) (equal bool) {
	return kc.Condition.Equal(another)
}

func (kc keyCondition) GetKey() string {
	return kc.Key
}
