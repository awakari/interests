package service

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/storage"
)

const (
	initialVersion = 0
)

var (

	// ErrConflict indicates the subscription exists in the underlying storage and can not be created.
	ErrConflict = errors.New("subscription already exists")

	// ErrNotFound indicates the subscription is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("subscription was not found")
)

type (

	// Service is a Subscription CRUDL service.
	Service interface {

		// Create a Subscription means subscribing.
		// Returns ErrConflict if a Subscription with the same name already present in the underlying storage.
		Create(ctx context.Context, sub model.Subscription) (err error)

		// Read the specified Subscription.
		// Returns ErrNotFound if Subscription is missing in the underlying storage.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Update replaces the existing Subscription.
		// Returns ErrNotFound if missing with the same Subscription.Name and Subscription.Version in the underlying storage.
		// In case of ErrNotFound it may be worth to Read the latest version again before Update retry.
		Update(ctx context.Context, sub model.Subscription) (err error)

		// Delete a Subscription means unsubscribing.
		// Returns ErrNotFound if Subscription with the specified Subscription.Name is missing in the underlying storage.
		Delete(ctx context.Context, name string) (err error)

		// Resolve all matching subscriptions by the specified message model.Metadata.
		// Once returns all the matching subscriptions should be available in the aggregator with the specified
		// model.MessageId. It's client responsibility to filter the model.Subscription candidates from the aggregator.
		Resolve(ctx context.Context, md model.MessageDescriptor) (err error)
	}

	service struct {
		s storage.Storage
	}
)

func NewService(s storage.Storage) Service {
	return service{
		s: s,
	}
}

func (svc service) Create(ctx context.Context, sub model.Subscription) (err error) {
	return nil
}

func (svc service) Read(ctx context.Context, name string) (sub model.Subscription, err error) {
	return model.Subscription{}, nil
}

func (svc service) Update(ctx context.Context, sub model.Subscription) (err error) {
	return nil
}

func (svc service) Delete(ctx context.Context, name string) (err error) {
	return nil
}

func (svc service) Resolve(ctx context.Context, md model.MessageDescriptor) (err error) {
	return nil
}
