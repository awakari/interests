package model

type (

	// Subscription represents the storage-level subscription entry.
	Subscription struct {
		SubscriptionKey

		// Description represents an optional human readable Subscription description.
		Description string

		// Includes represents a MatcherGroup to include the Subscription to query results.
		Includes MatcherGroup

		// Excludes represents a MatcherGroup to exclude the Subscription from the query results.
		Excludes MatcherGroup
	}
)
