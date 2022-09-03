package storage

import "github.com/meandros-messaging/subscriptions/service/patterns"

type (

	// Metadata is the incoming message metadata to match the subscriptions.
	Metadata map[string]string

	// MetadataPatternCodes is the map of the matched patterns by Metadata keys
	MetadataPatternCodes map[string][]patterns.PatternCode
)
