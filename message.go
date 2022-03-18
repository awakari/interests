package synapse

type Message struct {
	TopicNames []TopicName
	Data       interface{}
}
