package interest

// Interest represents the interest entry.
type Interest struct {

	// Id represents the unique Interest Id.
	Id string

	GroupId string

	// UserId represents an id of the Interest owner.
	UserId string

	// Data contains the Interest payload, mutable and immutable parts.
	Data Data
}
