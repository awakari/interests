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

func (sub Subscription) Matches(md Metadata, key string, patternCode PatternCode) (matches bool) {
	//includes := sub.Includes
	//if includes.All {
	//	for _, m := range includes.Matchers {
	//		// skip the matched before (key, patternCode) pair
	//		if key != m.Key || !bytes.Equal(patternCode, m.Pattern.Code) && !m.Matches(md) {
	//			matches = false
	//			break
	//		}
	//	}
	//}
	return
}
