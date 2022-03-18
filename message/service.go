package message

import "synapse"

// Service is a synapse.Message publishing service
type Service interface {

	// Send means Publish Event in Pub/Sub pattern
	Send(msg synapse.Message) error

	// SendBatch sends multiple messages at once
	SendBatch(msgBatch []synapse.Message) error
}
