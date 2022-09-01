package patterns

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	grpcApi "subscriptions/patterns/api/grpc"
)

type (

	// Code is a pattern identifier. Generally, not equal to the source pattern string.
	Code []byte

	// Service is CRUDL operations interface for the patterns.
	Service interface {

		// Create adds the specified Pattern if not present yet. Returns the created pattern Code, error otherwise.
		// May return ErrConcurrentUpdate, ErrInvalidPattern or ErrInternal when fails.
		Create(ctx context.Context, src string) (Code, error)

		// Read returns the Pattern by the specified Code if it exists. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Read(ctx context.Context, code Code) (string, error)

		// Delete removes the pattern if present by the specified Code. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Delete(ctx context.Context, code Code) error

		// SearchMatches finds all patterns Code matching the specified input.
		// Returns the page of results with count not more than the limit specified in the Query.
		// To search the next page use the last returned result from the previous page and set it to Query.Cursor.
		// Returns the empty results page ff search is complete (no more results).
		// May return ErrInternal when fails.
		SearchMatches(ctx context.Context, input string, limit uint32, cursor Code) ([]Code, error)
	}

	service struct {
		client grpcApi.ServiceClient
	}
)

var (

	// ErrConcurrentUpdate indicates there's a concurrent storage modification operation, safe to retry.
	ErrConcurrentUpdate = errors.New("concurrent modification, please retry")

	// ErrNotFound indicates there's no Pattern in the storage found by the specified Code.
	ErrNotFound = errors.New("pattern not found")

	// ErrInvalidPattern indicates the invalid pattern source input.
	ErrInvalidPattern = errors.New("invalid pattern input")

	// ErrInternal indicates there's internal failure. The best option to wrap ErrNodeDecode.
	ErrInternal = errors.New("internal failure")
)

func NewService(conn grpc.ClientConnInterface) Service {
	client := grpcApi.NewServiceClient(conn)
	return service{client: client}
}

func (svc service) Create(ctx context.Context, src string) (id Code, err error) {
	req := &grpcApi.CreateRequest{
		Src: src,
	}
	var resp *grpcApi.CreateResponse
	resp, err = svc.client.Create(ctx, req)
	if err == nil {
		id = resp.GetCode()
	}
	return
}

func (svc service) Read(ctx context.Context, code Code) (src string, err error) {
	req := &grpcApi.ReadRequest{
		Code: code,
	}
	var resp *grpcApi.ReadResponse
	resp, err = svc.client.Read(ctx, req)
	if err == nil {
		src = resp.GetSrc()
	}
	return
}

func (svc service) Delete(ctx context.Context, code Code) (err error) {
	req := &grpcApi.DeleteRequest{
		Code: code,
	}
	_, err = svc.client.Delete(ctx, req)
	return
}

func (svc service) SearchMatches(ctx context.Context, input string, limit uint32, cursor Code) (codes []Code, err error) {
	req := &grpcApi.SearchMatchesRequest{
		Input:  input,
		Limit:  limit,
		Cursor: cursor,
	}
	var resp *grpcApi.SearchMatchesResponse
	resp, err = svc.client.SearchMatches(ctx, req)
	if err == nil {
		results := resp.GetResults()
		for _, r := range results {
			codes = append(codes, r)
		}
	}
	return
}
