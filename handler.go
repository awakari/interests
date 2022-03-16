package synapse

type (

	// Handler is the message handler, it can be used to send messages:
	// * to SMS gateway
	// * via e-mail
	// * to the queue (Kafka, NATS, SQS, ...)
	// * ...
	Handler func(Message) error

	// HandlerConfig is the specific Handler configuration used by handler.Factory to init a specific Handler
	HandlerConfig map[string]interface{}
)
