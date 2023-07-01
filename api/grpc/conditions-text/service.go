package conditions_text

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	Create(ctx context.Context, k, v string, exact bool) (id, out string, err error)
	LockCreate(ctx context.Context, id string) (err error)
	UnlockCreate(ctx context.Context, id string) (err error)
	Delete(ctx context.Context, id string) (err error)
}

type service struct {
	client ServiceClient
}

var ErrInternal = errors.New("internal failure")

var ErrConflict = errors.New("already exists")

var ErrNotFound = errors.New("not found")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Create(ctx context.Context, k, v string, exact bool) (id, out string, err error) {
	req := CreateRequest{
		Key:   k,
		Term:  v,
		Exact: exact,
	}
	var resp *CreateResponse
	resp, err = svc.client.Create(ctx, &req)
	if err == nil {
		id = resp.Id
		out = resp.Term
	}
	err = decodeError(err)
	return
}

func (svc service) LockCreate(ctx context.Context, id string) (err error) {
	req := LockCreateRequest{
		Id: id,
	}
	_, err = svc.client.LockCreate(ctx, &req)
	err = decodeError(err)
	return
}

func (svc service) UnlockCreate(ctx context.Context, id string) (err error) {
	req := UnlockCreateRequest{
		Id: id,
	}
	_, err = svc.client.UnlockCreate(ctx, &req)
	err = decodeError(err)
	return
}

func (svc service) Delete(ctx context.Context, id string) (err error) {
	req := DeleteRequest{
		Id: id,
	}
	_, err = svc.client.Delete(ctx, &req)
	err = decodeError(err)
	return
}

func decodeError(src error) (dst error) {
	switch status.Code(src) {
	case codes.OK:
		dst = nil
	case codes.AlreadyExists:
		dst = fmt.Errorf("%w: %s", ErrConflict, src)
	case codes.NotFound:
		dst = fmt.Errorf("%w: %s", ErrNotFound, src)
	default:
		dst = fmt.Errorf("%w: %s", ErrInternal, src)
	}
	return
}
