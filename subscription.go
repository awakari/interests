package subscriptions

type (

	// Subscription connects source metadata filters with the destination message handler(s).
	Subscription struct {

		// Version is the subscription version used for optimistic update lock purpose.
		Version uint64

		Data
	}

	// Data contains the subscription details
	Data struct {

		// Name is the unique name of the Subscription.
		Name string

		// Description of the Subscription
		Description string

		// Matches contains the MetadataConstraint that should match the whole input string.
		Matches MetadataConstraint

		// ContainsMatches contains the MetadataConstraint that should match a part of the input string.
		ContainsMatches MetadataConstraint
	}

	// MetadataConstraint is the subscription filter for the incoming message metadata.
	MetadataConstraint struct {

		// Necessary contains the patterns those all should be present in the message metadata.
		Necessary MetadataPatterns

		// Sufficient contains the patterns those any should be present in the message metadata.
		Sufficient MetadataPatterns

		// Not contains the patterns those should not be present in the message metadata.
		Not MetadataPatterns
	}

	// MetadataPatterns is the map of patterns by the key, e.g. "subject": ["foo*", "?bar"].
	MetadataPatterns map[string][]string
)
