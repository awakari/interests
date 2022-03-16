package message

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"synapse"
	"synapse/handler"
	"synapse/subscription"
	"synapse/topic"
	"synapse/util"
	"sync"
)

const (
	SubscriptionsQueryLimit = 100
)

var (
	log = logrus.New()

	ErrTopicNotSet = errors.New("message topic not set")
)

type (

	// Service is a synapse.Message publishing service
	Service interface {

		// Send means Publish Event in Pub/Sub pattern. Returns an assigned random synapse.MessageId.
		// Returns error when failed to deliver to any of the associated destinations.
		// Returned non-nil error doesn't mean the message has not been delivered to another destination (if multiple).
		Send(msg synapse.MessageData) (synapse.MessageId, error)
	}

	service struct {
		handlerSvc      handler.Service
		subscriptionSvc subscription.Service
		topicSvc        topic.Service
	}
)

func NewService(handlerSvc handler.Service, subscriptionSvc subscription.Service, topicSvc topic.Service) Service {
	return service{
		handlerSvc:      handlerSvc,
		subscriptionSvc: subscriptionSvc,
		topicSvc:        topicSvc,
	}
}

func (svc service) Send(md synapse.MessageData) (synapse.MessageId, error) {

	id := synapse.MessageId(rand.Uint64())
	msg := synapse.Message{
		Id:          id,
		MessageData: md,
	}
	log.Debug(fmt.Sprintf("Incoming message %v", msg))

	if msg.RequestTopicName == "" {
		return id, ErrTopicNotSet
	}
	t, err := svc.topicSvc.Read(msg.RequestTopicName)
	if err != nil {
		return id, err
	}
	log.Debug(fmt.Sprintf("Resolved the request topic \"%s\" for the message id %d", msg.RequestTopicName, msg.Id))

	wg := sync.WaitGroup{}
	errChan := make(chan error, 1) // keep the 1st error only
	var cursorRef *synapse.SubscriptionsPageCursor = nil
	for {
		query := synapse.SubscriptionsQuery{
			TopicIdRef: &t.Id,
			PageQuery: util.PageQuery[synapse.SubscriptionsPageCursor]{
				CursorRef: cursorRef,
				Limit:     SubscriptionsQueryLimit,
			},
		}
		page, err := svc.subscriptionSvc.List(query)
		if err != nil {
			return id, err
		}
		if len(page.Items) > 0 {
			wg.Add(1)
			go svc.sendToSubscriptions(msg, page.Items, &wg, errChan)
		}
		if page.Complete {
			break
		}
		cursorRef = page.CursorRef
	}
	wg.Wait()
	close(errChan)

	// don't block if errors channel doesn't contain anything
	select {
	case err := <-errChan:
		return id, err
	default:
		return id, nil
	}
}

func (svc service) sendToSubscriptions(
	msg synapse.Message,
	ss []synapse.Subscription,
	parentWgRef *sync.WaitGroup,
	errChan chan<- error,
) {

	defer parentWgRef.Done()

	handlers, err := svc.handlerSvc.Resolve(ss)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to resolve handlers by subscriptions: %v, %s", ss, err))
		// don't block if there's another error in the errors channel already
		select {
		case errChan <- err:
		default:
		}
		return
	}

	wg := sync.WaitGroup{}
	for _, h := range handlers {
		wg.Add(1)
		go invokeHandler(h, msg, &wg, errChan)
	}
	wg.Wait()
}

func invokeHandler(h synapse.Handler, msg synapse.Message, parentWgRef *sync.WaitGroup, errChan chan<- error) {

	defer parentWgRef.Done()

	err := h(msg)
	if err != nil {
		log.Error(fmt.Sprintf("Handler %v failed to process the message id %d: %s", h, msg.Id, err))
		// don't block if there's another error in the errors channel already
		select {
		case errChan <- err:
		default:
		}
	}
}
