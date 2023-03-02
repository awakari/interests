package kiwiTree

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (

	// Service represents the kiwi-tree service.
	Service interface {

		// Create creates a new key/pattern pair and registers it. Repeating the same Create call yields the same state.
		// Returns ErrShouldRetry when the corresponding entity in the underlying storage is temporarily locked.
		Create(ctx context.Context, k string, pattern string) (err error)

		// LockCreate sets a lock in the underlying storage to prevent the creation of the same key/pattern pair.
		// Returns ErrNotFound if the specified pattern is not present.
		LockCreate(ctx context.Context, k string, pattern string) (err error)

		// UnlockCreate unsets the lock set by LockCreate, if any. Otherwise, the result depends on the underlying
		// storage implementation.
		UnlockCreate(ctx context.Context, k string, pattern string) (err error)

		// Delete removes a specified key/pattern from the underlying storage.
		// Returns ErrNotFound when the specified key/pattern pair is missing in the underlying storage.
		Delete(ctx context.Context, k string, pattern string) (err error)
	}

	service struct {
		client ServiceClient
	}
)

var (

	// ErrInvalidPatternSrc indicates the source string to create a Pattern is invalid.
	ErrInvalidPatternSrc = errors.New("invalid pattern source")

	// ErrInternal indicates some unexpected internal failure.
	ErrInternal = errors.New("internal failure")

	// ErrNotFound indicates the specified key/pattern pair was not found in the underlying storage.
	ErrNotFound = errors.New("not found")

	// ErrShouldRetry indicates a storage entity is locked and the operation should be retried.
	ErrShouldRetry = errors.New("retry the operation")
)

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Create(ctx context.Context, k string, pattern string) (err error) {
	req := &KeyPatternRequest{
		Key:     k,
		Pattern: pattern,
	}
	_, err = svc.client.Create(ctx, req)
	if err != nil {
		err = decodeError(err)
	}
	return
}

func (svc service) LockCreate(ctx context.Context, k string, pattern string) (err error) {
	req := &KeyPatternRequest{
		Key:     k,
		Pattern: pattern,
	}
	_, err = svc.client.LockCreate(ctx, req)
	if err != nil {
		err = decodeError(err)
	}
	return err
}

func (svc service) UnlockCreate(ctx context.Context, k string, pattern string) (err error) {
	req := &KeyPatternRequest{
		Key:     k,
		Pattern: pattern,
	}
	_, err = svc.client.UnlockCreate(ctx, req)
	if err != nil {
		err = decodeError(err)
	}
	return err
}

func (svc service) Delete(ctx context.Context, k string, pattern string) (err error) {
	req := &KeyPatternRequest{
		Key:     k,
		Pattern: pattern,
	}
	_, err = svc.client.Delete(ctx, req)
	if err != nil {
		err = decodeError(err)
	}
	return err
}

func decodeError(grpcErr error) (svcErr error) {
	switch status.Code(grpcErr) {
	case codes.OK:
		svcErr = nil
	case codes.InvalidArgument:
		svcErr = fmt.Errorf("%w: %s", ErrInvalidPatternSrc, grpcErr)
	case codes.NotFound:
		svcErr = fmt.Errorf("%w: %s", ErrNotFound, grpcErr)
	case codes.Unavailable:
		svcErr = fmt.Errorf("%w: %s", ErrShouldRetry, grpcErr)
	default:
		svcErr = fmt.Errorf("%w: %s", ErrInternal, grpcErr)
	}
	return
}
