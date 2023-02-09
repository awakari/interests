package condition

type (
	KeyCondition interface {
		Condition
		GetId() string
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
	equal = kc.Condition.Equal(another)
	if equal {
		var anotherKc KeyCondition
		anotherKc, equal = another.(KeyCondition)
		if equal {
			equal = kc.Key == anotherKc.GetKey()
		}
	}
	return
}

func (kc keyCondition) GetKey() string {
	return kc.Key
}
