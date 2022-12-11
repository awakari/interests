package model

type (
	MetadataRule interface {
		Rule
		GetKey() string
	}

	metadataRule struct {
		Rule Rule
		Key  string
	}
)

func NewMetadataRule(r Rule, k string) MetadataRule {
	return metadataRule{
		Rule: r,
		Key:  k,
	}
}

func (mr metadataRule) GetKey() string {
	return mr.Key
}

func (mr metadataRule) IsNot() bool {
	return mr.Rule.IsNot()
}

func (mr metadataRule) Equal(another Rule) (equal bool) {
	equal = mr.Rule.Equal(another)
	if equal {
		var anotherMr MetadataRule
		anotherMr, equal = another.(MetadataRule)
		if equal {
			equal = mr.Key == anotherMr.GetKey()
		}
	}
	return
}
