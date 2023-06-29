package condition

type (
	KeyCondition interface {
		Condition
		GetId() string
		SetId(id string)
		GetKey() string
	}

	keyCondition struct {
		Condition Condition
		Id        string
		Key       string
	}
)

func NewKeyCondition(c Condition, id, k string) KeyCondition {
	return &keyCondition{
		Condition: c,
		Id:        id,
		Key:       k,
	}
}

func (kc *keyCondition) IsNot() bool {
	return kc.Condition.IsNot()
}

func (kc *keyCondition) GetId() string {
	return kc.Id
}

func (kc *keyCondition) SetId(id string) {
	kc.Id = id
}

func (kc *keyCondition) Equal(another Condition) (equal bool) {
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

func (kc *keyCondition) GetKey() string {
	return kc.Key
}
