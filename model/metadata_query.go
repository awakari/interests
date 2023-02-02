package model

type MetadataQuery struct {
	Limit uint32

	Metadata map[string]string
}
