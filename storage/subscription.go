package storage

type (
	Subscription struct {
		Name            string
		Version         uint64
		Description     string
		Matches         MetadataConstraint
		ContainsMatches MetadataConstraint
	}

	MetadataConstraint struct {
		Required   MetadataPatterns
		Sufficient MetadataPatterns
		Excluding  MetadataPatterns
	}

	MetadataPatterns map[string][]PatternId

	PatternId []byte
)
