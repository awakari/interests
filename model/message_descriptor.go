package model

type (
	MessageId uint64

	MessageDescriptor struct {
		Id       MessageId
		Metadata Metadata
	}
)
