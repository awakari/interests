package model

type (
	MetadataPatternRule interface {
		MetadataRule
		IsPartial() bool
		GetPattern() Pattern
	}

	metadataPatternRule struct {
		MetadataRule MetadataRule
		Partial      bool
		Pattern      Pattern
	}
)

func NewMetadataPatternRule(mr MetadataRule, partial bool, p Pattern) MetadataPatternRule {
	return metadataPatternRule{
		MetadataRule: mr,
		Partial:      partial,
		Pattern:      p,
	}
}

func (mpr metadataPatternRule) IsNot() bool {
	return mpr.MetadataRule.IsNot()
}

func (mpr metadataPatternRule) Equal(another Rule) (equal bool) {
	equal = mpr.MetadataRule.Equal(another)
	if equal {
		var anotherMpr MetadataPatternRule
		anotherMpr, equal = another.(MetadataPatternRule)
		if equal {
			equal = mpr.Partial == anotherMpr.IsPartial() && mpr.Pattern.Equal(anotherMpr.GetPattern())
		}
	}
	return
}

func (mpr metadataPatternRule) GetKey() string {
	return mpr.MetadataRule.GetKey()
}

func (mpr metadataPatternRule) IsPartial() bool {
	return mpr.Partial
}

func (mpr metadataPatternRule) GetPattern() Pattern {
	return mpr.Pattern
}
