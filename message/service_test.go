package message

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"synapse"
	"synapse/handler"
	"synapse/handler/factory"
	"synapse/subscription"
	"synapse/topic"
	"testing"
)

func TestSend(t *testing.T) {

	msgChan := make(chan synapse.Message, 2)
	msgChanRef := &msgChan
	f := factory.NewFactoryMock(
		func(msg synapse.Message) error {
			*msgChanRef <- msg
			return nil
		},
	)

	factoryStorage := make(map[string]factory.Factory)
	factorySvc := factory.NewService(factoryStorage)
	err := factorySvc.Register("factory0", f)
	require.Nil(t, err)

	handlerStorage := make(map[synapse.SubscriptionId]synapse.Handler)
	handlerSvc := handler.NewServiceMock(factorySvc, handlerStorage)

	subscriptionStorage := make(map[synapse.SubscriptionId]synapse.SubscriptionData)
	subscriptionSvc := subscription.NewServiceMock(subscriptionStorage)
	_, err = subscriptionSvc.Create(
		synapse.SubscriptionData{
			Description:        "subscription0",
			TopicIds:           []synapse.TopicId{"topic0"},
			HandlerFactoryName: "factory0",
			HandlerConfig:      nil,
		},
	)
	require.Nil(t, err)

	topicStorage := make(map[synapse.TopicId]synapse.TopicData)
	topicSvc := topic.NewServiceMock(topicStorage, subscriptionSvc)
	topic0 := synapse.Topic{
		Id: "topic0",
		TopicData: synapse.TopicData{
			Name:        "topic0",
			Description: "test topic 0",
		},
	}
	_, err = topicSvc.Create(topic0.TopicData)
	require.Nil(t, err)

	svc := NewService(handlerSvc, subscriptionSvc, topicSvc)

	msgData := synapse.MessageData{
		RequestTopicName:  "topic0",
		ResponseTopicName: "topic1",
		Payload:           []byte{3, 141, 59, 26},
	}
	msgId0, err := svc.Send(msgData)
	assert.Nil(t, err)
	msgId1, err := svc.Send(msgData)
	assert.Nil(t, err)
	close(msgChan)
	assert.NotEqual(t, msgId0, msgId1)

	deliveryCount := 0
	for msg := range msgChan {
		msgId := msg.Id
		if msgId != msgId0 {
			assert.Equal(t, msgId1, msgId)
		}
		assert.Equal(t, "topic0", msg.RequestTopicName)
		assert.Equal(t, "topic1", msg.ResponseTopicName)
		assert.Equal(t, []byte{3, 141, 59, 26}, msg.Payload)
		deliveryCount++
	}
	assert.Equal(t, 2, deliveryCount)
}
