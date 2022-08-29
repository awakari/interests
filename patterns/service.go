package patterns

import (
	"context"
	"google.golang.org/grpc"
	grpcApi "subscriptions/patterns/api/grpc"
)

type (

	// Id is a pattern identifier. Generally, not equal to the source pattern string.
	Id []byte

	// Service is CRUDL operations interface for the patterns.
	Service interface {

		// Create adds the specified Pattern if not present yet. Returns the created pattern Id, error otherwise.
		// May return ErrConcurrentUpdate, ErrInvalidPattern or ErrInternal when fails.
		Create(ctx context.Context, src string) (Id, error)

		// Read returns the Pattern by the specified Id if it exists. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Read(ctx context.Context, id Id) (string, error)

		// Delete removes the pattern if present by the specified Id. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Delete(ctx context.Context, id Id) error

		// SearchMatches finds all patterns Id matching the specified input.
		// Returns the page of results with count not more than the limit specified in the Query.
		// To search the next page use the last returned result from the previous page and set it to Query.Cursor.
		// Returns the empty results page ff search is complete (no more results).
		// May return ErrInternal when fails.
		SearchMatches(ctx context.Context, input string, limit uint32, cursor Id) ([]Id, error)
	}

	service struct {
		client grpcApi.ServiceClient
	}
)

func NewService(conn grpc.ClientConnInterface) Service {
	client := grpcApi.NewServiceClient(conn)
	return service{client: client}
}

func (svc service) Create(ctx context.Context, src string) (id Id, err error) {
	req := &grpcApi.CreateRequest{
		Src: src,
	}
	var resp *grpcApi.CreateResponse
	resp, err = svc.client.Create(ctx, req)
	if err == nil {
		id = resp.GetPath()
	}
	return
}

func (svc service) Read(ctx context.Context, id Id) (src string, err error) {
	req := &grpcApi.ReadRequest{
		Path: id,
	}
	var resp *grpcApi.ReadResponse
	resp, err = svc.client.Read(ctx, req)
	if err == nil {
		src = resp.GetSrc()
	}
	return
}

func (svc service) Delete(ctx context.Context, id Id) (err error) {
	req := &grpcApi.DeleteRequest{
		Path: id,
	}
	_, err = svc.client.Delete(ctx, req)
	return
}

func (svc service) SearchMatches(ctx context.Context, input string, limit uint32, cursor Id) (ids []Id, err error) {
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
			ids = append(ids, r)
		}
	}
	return
}
