package handler

import (
	"synapse/message"
)

type (

	// Handler is the message handler, it can be used to send messages:
	// * to SMS gateway
	// * via e-mail
	// * to the queue (Kafka, NATS, SQS, ...)
	// * ...
	Handler func(message.Message) error

	// Config is the specific Handler configuration used by handler.Factory to init a specific Handler
	Config map[string]interface{}
)
