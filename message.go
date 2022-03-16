package synapse

type (

	// Message is the basic entity containing the incoming id, destination topic and any Data
	Message struct {

		// Id is the message incoming id
		Id MessageId

		MessageData
	}

	// MessageId is the likely unique Message id for the logging/tracing purposes
	MessageId uint64

	// MessageData is the message data not identified yet
	MessageData struct {

		// RequestTopicName is the destination topic name
		RequestTopicName string

		// ResponseTopicName is the optional topic name for the response. Empty string means not set.
		ResponseTopicName string

		// Payload is the serialized payload data
		Payload []byte
	}
)
