package topic

import (
	"errors"
	"synapse"
)

var (
	ErrConflict = errors.New("topic already exists")

	ErrNotFound = errors.New("topic was not found")
)

type (

	// Service is the synapse.Topic service
	Service interface {

		// Create a new synapse.Topic
		Create(data synapse.TopicData) (synapse.TopicId, error)

		// Read the synapse.Topic by its unique name
		Read(name string) (*synapse.Topic, error)

		// Delete a topic by its unique name. Also deletes all synapse.Subscription to this synapse.Topic.
		Delete(name string) error

		// List all known topics with the pagination support
		List(query synapse.TopicsQuery) (synapse.TopicsPage, error)
	}
)
