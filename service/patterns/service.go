package patterns

import (
	"context"
	"errors"
	"fmt"
	apiGrpc "github.com/meandros-messaging/subscriptions/api/grpc/patterns"
	"github.com/meandros-messaging/subscriptions/storage"
	"google.golang.org/grpc"
)

type (

	// BulkCursor represents the bulk search matches cursor.
	BulkCursor struct {

		// Key is the last metadata key cursor.
		Key string

		// PatternCode is the last pattern Code.
		PatternCode PatternCode
	}

	// Service is CRUDL operations interface for the patterns.
	Service interface {

		// Create adds the specified Pattern if not present yet. Returns the created pattern Code, error otherwise.
		// May return ErrConcurrentUpdate, ErrInvalidPattern or ErrInternal when fails.
		Create(ctx context.Context, src string) (PatternCode, error)

		// Read returns the Pattern by the specified Code if it exists. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Read(ctx context.Context, code PatternCode) (Pattern, error)

		// Delete removes the pattern if present by the specified Code. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Delete(ctx context.Context, code PatternCode) error

		// SearchMatchesBulk finds all patterns Code matching any of the specified metadata values.
		// This is the convenience bulk function returning the codes grouped by the input metadata keys.
		// A client should use the greatest key and last pattern Code from the previous results page to provide a next cursor.
		// Returns the page of results with count not more than the limit specified in the query.
		SearchMatchesBulk(ctx context.Context, md storage.Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Pattern, err error)
	}

	service struct {
		client apiGrpc.ServiceClient
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
	client := apiGrpc.NewServiceClient(conn)
	return service{client: client}
}

func (svc service) Create(ctx context.Context, src string) (id PatternCode, err error) {
	req := &apiGrpc.CreateRequest{
		Src: src,
	}
	var resp *apiGrpc.CreateResponse
	resp, err = svc.client.Create(ctx, req)
	if err == nil {
		id = resp.GetCode()
	}
	return
}

func (svc service) Read(ctx context.Context, code PatternCode) (p Pattern, err error) {
	req := &apiGrpc.ReadRequest{
		Code: code,
	}
	var resp *apiGrpc.ReadResponse
	resp, err = svc.client.Read(ctx, req)
	if err == nil {
		respPattern := resp.Pattern
		if respPattern == nil {
			err = fmt.Errorf("%w by code: %s", ErrNotFound, code)
		} else {
			p = Pattern{
				Code:  respPattern.Code,
				Regex: respPattern.Regex,
				Src:   respPattern.Src,
			}
		}
	}
	return
}

func (svc service) Delete(ctx context.Context, code PatternCode) (err error) {
	req := &apiGrpc.DeleteRequest{
		Code: code,
	}
	_, err = svc.client.Delete(ctx, req)
	return
}

func (svc service) SearchMatchesBulk(ctx context.Context, md storage.Metadata, limit uint32, cursor *BulkCursor) (page map[string][]Pattern, err error) {
	var reqCursor *apiGrpc.BulkCursor = nil
	if cursor != nil {
		reqCursor = &apiGrpc.BulkCursor{
			Key:         cursor.Key,
			PatternCode: cursor.PatternCode,
		}
	}
	req := &apiGrpc.SearchMatchesBulkRequest{
		Md:     md,
		Limit:  limit,
		Cursor: reqCursor,
	}
	var resp *apiGrpc.SearchMatchesBulkResponse
	resp, err = svc.client.SearchMatchesBulk(ctx, req)
	if err == nil {
		results := resp.Results
		page = make(map[string][]Pattern, len(results))
		for k, respPatterns := range results {
			var ps []Pattern
			for _, respPattern := range respPatterns.Values {
				p := Pattern{
					Code:  respPattern.Code,
					Regex: respPattern.Regex,
					Src:   respPattern.Src,
				}
				ps = append(ps, p)
			}
			page[k] = ps
		}
	}
	return
}
