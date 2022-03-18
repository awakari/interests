package util

// ResultsPage is the generic query results pagination data type where:
// * R is a single result type
// * C is the pagination cursor type
type ResultsPage[R interface{}, C interface{}] struct {

	// Results is the current ResultsPage result slice
	Results []R

	// Complete is true if reached end of results, false otherwise
	Complete bool

	// Cursor is the next ResultsPage cursor
	Cursor C
}
