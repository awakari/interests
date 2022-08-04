package message

import (
	"fmt"
	"math/rand"
)

type (

	// ErrNoBuffSpace means that a Message could not be put into the buffer without blocking
	ErrNoBuffSpace error

	// Service is a message.Message publishing service
	Service interface {

		// Send puts the incoming message to the output buffer.
		// Returns a new random message.Id what should be unique in the big time window.
		// Returns ErrNoBuffSpace when unable to put a message into the output buffer without blocking.
		Send(data Data) (Id, error)
	}

	service struct {
		buffRef *chan<- Message
	}
)

func NewService(buffRef *chan<- Message) Service {
	return service{
		buffRef: buffRef,
	}
}

func (svc service) Send(data Data) (Id, error) {
	id := Id(rand.Uint64())
	msg := Message{
		Id:   id,
		Data: data,
	}
	select {
	case *svc.buffRef <- msg:
	default:
		return id, fmt.Errorf("send message %v failed: no space in the buffer", msg).(ErrNoBuffSpace)
	}
	return id, nil
}
