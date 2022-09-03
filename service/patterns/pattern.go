package patterns

type (

	// PatternCode is a pattern identifier. Generally, not equal to the source pattern string.
	PatternCode []byte

	Pattern struct {
		Code  []byte
		Regex string
		Src   string
	}
)
