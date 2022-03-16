package synapse

import "synapse/util"

type (

	// Topic is the Message channel. Any Message consumer can subscribe for one or many topics to receive the Message.
	Topic struct {
		Id TopicId

		TopicData
	}

	TopicId string

	TopicData struct {

		// Name is the unique topic name
		Name string

		// Description is the free form Topic description
		Description string
	}

	TopicName string

	TopicsPageCursor TopicId

	TopicsPage struct {
		util.Page[Topic, TopicsPageCursor]
	}

	TopicsQuery struct {
		NamePrefix string
		util.PageQuery[TopicsPageCursor]
	}
)
