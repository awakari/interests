package message

type (

	// Message is the basic entity containing the incoming id, destination topic and any Data
	Message struct {

		// Id is the message incoming id
		Id Id

		Data
	}

	// Id is the likely unique Message id for the logging/tracing purposes
	Id uint64

	// Data is the message data not identified yet

	Data struct {

		// Metadata is free-form metadata tree.
		Metadata map[string]interface{}

		// Payload is the serialized payload data
		Payload []byte
	}
)
