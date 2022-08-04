package message

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSend(t *testing.T) {

	msgChan := make(chan Message, 3)
	outBuff := chan<- Message(msgChan)
	svc := NewService(&outBuff)

	msgData0 := Data{
		Metadata: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": 42,
			},
		},
		Payload: []byte{31, 41, 59, 26},
	}
	//msgData1 := Data{
	//	Metadata: map[string]interface{}{
	//		"subject": "updates",
	//	},
	//	Payload: []byte{27, 182, 81, 82},
	//}

	cases := map[string]struct {
		send      []Data
		errors    []error
		delivered []Data
	}{
		"Should send same message multiple times successfully": {
			send: []Data{
				msgData0,
				msgData0,
			},
			errors: []error{
				nil,
				nil,
			},
			delivered: []Data{
				msgData0,
				msgData0,
			},
		},
	}

	for desc, tc := range cases {
		msgIds := make([]Id, len(tc.send))
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
			assert.Equal(t, expectedDelivery, msg.Data)
			assert.Equal(t, expectedDelivery, msg.Data)
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
