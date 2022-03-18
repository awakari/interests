package synapse

import "synapse/util"

type TopicId int

type TopicName string

type TopicData struct {
	Name        TopicName
	Description string
}

type Topic struct {
	Id TopicId
	TopicData
}

type TopicsPageCursor int

type TopicsPage struct {
	util.ResultsPage[TopicData, TopicsPageCursor]
}
