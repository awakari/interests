package subscription

// Data represents the Subscription payload data.
type Data struct {

	// Metadata represents the optional subscription attributes, e.g. human-readable description, user ownership.
	Metadata map[string]string

	// Route represent the Subscription routing data.
	Route Route
}
