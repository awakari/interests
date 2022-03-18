package synapse

import (
	"synapse/util"
)

// Handler is the message handler, it can be used to send messages:
// * to SMS gateway
// * via e-mail
// * to the queue (Kafka, NATS, SQS, ...)
// * ...
type Handler struct {

	// Name is the unique Handler name
	Name string

	// Description is the Handler description
	Description string

	// HandleFunc is a generic Message processing function
	HandleFunc util.HandleFunc[Message]

	// HandleBatchFunc is a generic batch Message processing function.
	// Returns count of processed events if not all processed at once.
	HandleBatchFunc util.HandleBatchFunc[Message]
}
