package model

type (

	// Subscription represents the storage-level subscription entry.
	Subscription struct {

		// Name represents a unique subscription name.
		Name string

		// Version represents a Subscription entry version for the optimistic lock purpose.
		Version uint64

		// Description represents an optional human readable Subscription description.
		Description string

		// Includes represents a MatcherGroup to include the Subscription to query results.
		Includes MatcherGroup

		// Excludes represents a MatcherGroup to exclude the Subscription from the query results.
		Excludes MatcherGroup
	}
)

func (sub Subscription) Matches(md Metadata, key string, patternCode PatternCode) (matches bool, err error) {
	matches, err = sub.Excludes.Matches(md, key, patternCode)
	if err == nil && !matches { // excludes group should not match
		matches, err = sub.Includes.Matches(md, key, patternCode)
	}
	return
}
