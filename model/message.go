package model

type (
	MessageId uint64

	Message struct {
		Id       MessageId
		Metadata Metadata
		Data     []byte
	}
)
