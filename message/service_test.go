package message

import (
	"errors"
	"fmt"
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

	msgChan := make(chan synapse.Message, 3)
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
	_, err = subscriptionSvc.Create(
		synapse.SubscriptionData{
			Description:        "subscription1",
			TopicIds:           []synapse.TopicId{"topic1"},
			HandlerFactoryName: "factory0",
			HandlerConfig:      nil,
		},
	)
	require.Nil(t, err)
	_, err = subscriptionSvc.Create(
		synapse.SubscriptionData{
			Description:        "subscription2",
			TopicIds:           []synapse.TopicId{"topic1"},
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
	topic1 := synapse.Topic{
		Id: "topic1",
		TopicData: synapse.TopicData{
			Name:        "topic1",
			Description: "test topic 1",
		},
	}
	_, err = topicSvc.Create(topic1.TopicData)
	require.Nil(t, err)
	topic2 := synapse.Topic{
		Id: "topic2",
		TopicData: synapse.TopicData{
			Name:        "topic2",
			Description: "test topic 2 - no subscription",
		},
	}
	_, err = topicSvc.Create(topic2.TopicData)
	require.Nil(t, err)

	svc := NewService(handlerSvc, subscriptionSvc, topicSvc)

	msgDataWithRequestTopicOnly := synapse.MessageData{
		RequestTopicName:  "topic0",
		ResponseTopicName: "topic3",
		Payload:           []byte{31, 41, 59, 26},
	}
	msgDataWithResponseTopicOnly := synapse.MessageData{
		RequestTopicName:  "",
		ResponseTopicName: "topic3",
		Payload:           []byte{27, 182, 81, 82},
	}
	msgDataWithMissingRequestTopicOnly := synapse.MessageData{
		RequestTopicName:  "topic3",
		ResponseTopicName: "",
		Payload:           []byte{31, 41, 59, 26},
	}
	msgDataWithBothTopics := synapse.MessageData{
		RequestTopicName:  "topic0",
		ResponseTopicName: "topic3",
		Payload:           []byte{27, 182, 81, 82},
	}
	msgDataWithNoSubscription := synapse.MessageData{
		RequestTopicName:  "topic2",
		ResponseTopicName: "",
		Payload:           []byte{31, 41, 59, 26},
	}
	msgDataWithMultipleSubscription := synapse.MessageData{
		RequestTopicName:  "topic1", // has multiple subscriptions
		ResponseTopicName: "",
		Payload:           []byte{27, 182, 81, 82},
	}

	cases := map[string]struct {
		send      []synapse.MessageData
		errors    []error
		delivered []synapse.MessageData
	}{
		"Should send same message multiple times successfully": {
			send: []synapse.MessageData{
				msgDataWithRequestTopicOnly,
				msgDataWithRequestTopicOnly,
				msgDataWithRequestTopicOnly,
			},
			errors: []error{
				nil,
				nil,
				nil,
			},
			delivered: []synapse.MessageData{
				msgDataWithRequestTopicOnly,
				msgDataWithRequestTopicOnly,
				msgDataWithRequestTopicOnly,
			},
		},
		"Should fail sending a message without request topic set": {
			send: []synapse.MessageData{
				msgDataWithResponseTopicOnly,
			},
			errors: []error{
				ErrTopicNotSet,
			},
			delivered: []synapse.MessageData{},
		},
		"Should fail sending a message to a missing topic": {
			send: []synapse.MessageData{
				msgDataWithMissingRequestTopicOnly,
			},
			errors: []error{
				topic.ErrNotFound,
			},
			delivered: []synapse.MessageData{},
		},
		"Should send successfully if response topic is missing": {
			send: []synapse.MessageData{
				msgDataWithBothTopics,
			},
			errors: []error{
				nil,
			},
			delivered: []synapse.MessageData{
				msgDataWithBothTopics,
			},
		},
		"Should send successfully when no subscription": {
			send: []synapse.MessageData{
				msgDataWithNoSubscription,
			},
			errors: []error{
				nil,
			},
			delivered: []synapse.MessageData{},
		},
		"Should send successfully to many subscriptions": {
			send: []synapse.MessageData{
				msgDataWithMultipleSubscription,
			},
			errors: []error{
				nil,
			},
			delivered: []synapse.MessageData{
				msgDataWithMultipleSubscription,
				msgDataWithMultipleSubscription,
			},
		},
	}

	for desc, tc := range cases {
		msgIds := make([]synapse.MessageId, len(tc.send))
		for i, msgData := range tc.send {
			msgId, err := svc.Send(msgData)
			assert.Equal(t, tc.errors[i], err, fmt.Sprintf("%s: expected error %s but got %s", desc, tc.errors[i], err))
			for j, prevMsgId := range msgIds {
				assert.NotEqual(t, prevMsgId, msgId, fmt.Sprintf("%s: message id %d equals to one of the previous #%d", desc, prevMsgId, j))
			}
			msgIds = append(msgIds, msgId)
		}
		for _, expectedDelivery := range tc.delivered {
			msg := <-msgChan
			assert.Equal(t, expectedDelivery.RequestTopicName, msg.RequestTopicName)
			assert.Equal(t, expectedDelivery.ResponseTopicName, msg.ResponseTopicName)
			assert.ElementsMatch(t, expectedDelivery.Payload, msg.Payload)
		}
	}

	select {
	case unexpectedMsg := <-msgChan:
		assert.Error(t, errors.New(fmt.Sprintf("Unexpected extra message delivered: %v", unexpectedMsg)))
	default:
	}

	close(msgChan)
}
