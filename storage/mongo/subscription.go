package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	patternId []byte

	subsription struct {
		IntId           primitive.ObjectID `bson:"_id"`
		Name            string
		Description     string
		Matches         metadataConstraint
		ContainsMatches metadataConstraint `bson:"contains_matches"`
	}

	metadataConstraint struct {
		Necessary  metadataPatterns
		Sufficient metadataPatterns
		Not        metadataPatterns
	}

	metadataPatterns map[string][]patternId
)

const (
	attrIntId           = "_id"
	attrExtId           = "ext_id"
	attrDescription     = "description"
	attrMatches         = "matches"
	attrContainsMatches = "contains_matches"
	attrNecessary       = "necessary"
	attrSufficient      = "sufficient"
	attrNot             = "not"
)
