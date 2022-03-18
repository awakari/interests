package topic

import "synapse"

type Service interface {

	// Create a new synapse.Topic
	Create(data synapse.TopicData) error

	// Read the synapse.TopicData by the synapse.TopicName
	Read(name synapse.TopicName) (synapse.TopicData, error)

	// Update the synapse.TopicData
	Update(data synapse.TopicData) error

	// Delete a topic by its synapse.TopicName
	Delete(name synapse.TopicName) error

	// List all known topics with the pagination support
	List(cursor synapse.TopicsPageCursor) (synapse.TopicsPage, error)
}
