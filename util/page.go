package util

type (

	// Page is the generic pagination data type where:
	// * I is am item type
	// * C is the pagination cursor type
	Page[I interface{}, C interface{}] struct {

		// Items is the current Page result slice
		Items []I

		// Complete is true if reached end of results, false otherwise
		Complete bool

		// CursorRef is the next Page cursor, may be nil (e.g. if there's no results)
		CursorRef *C
	}

	// PageQuery is a query for the results Page with a specified
	PageQuery[C interface{}] struct {

		// CursorRef is the reference to the optional cursor. May be nil if not present (query from the beginning).
		CursorRef *C

		// Limit query result count
		Limit uint
	}
)
