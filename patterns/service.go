package patterns

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	grpcApi "subscriptions/api/grpc/patterns"
)

type (

	// Code is a pattern identifier. Generally, not equal to the source pattern string.
	Code []byte

	Metadata map[string]string

	// BulkCursor represents the bulk search matches cursor.
	BulkCursor struct {

		// Key is the last metadata key cursor.
		Key string

		// PatternCode is the last pattern Code.
		PatternCode Code
	}

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

		// SearchMatchesBulk finds all patterns Code matching any of the specified metadata values.
		// This is the convenience bulk function returning the codes grouped by the input metadata keys.
		// A client should use the greatest key and last pattern Code from the previous results page to provide a next cursor.
		// Returns the page of results with count not more than the limit specified in the query.
		SearchMatchesBulk(ctx context.Context, md Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Code, err error)
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

func (svc service) SearchMatchesBulk(ctx context.Context, md Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Code, err error) {
	var reqCursor *grpcApi.BulkCursor = nil
	if cursor != nil {
		reqCursor = &grpcApi.BulkCursor{
			Key:         cursor.Key,
			PatternCode: cursor.PatternCode,
		}
	}
	req := &grpcApi.SearchMatchesBulkRequest{
		Md:     md,
		Limit:  limit,
		Cursor: reqCursor,
	}
	var resp *grpcApi.SearchMatchesBulkResponse
	resp, err = svc.client.SearchMatchesBulk(ctx, req)
	if err == nil {
		results := resp.Results
		page = make(map[string][]Code, len(results))
		for k, respCodes := range results {
			var codes []Code
			for _, c := range respCodes.Value {
				codes = append(codes, c)
			}
			page[k] = codes
		}
	}
	return
}
