package patterns

import (
	"context"
	"errors"
	"fmt"
	apiGrpc "github.com/meandros-messaging/subscriptions/api/grpc/patterns"
	"github.com/meandros-messaging/subscriptions/model"
	"google.golang.org/grpc"
)

type (

	// BulkCursor represents the bulk search matches cursor.
	BulkCursor struct {

		// Key is the last metadata key cursor.
		Key string

		// PatternCode is the last pattern Code.
		PatternCode model.PatternCode
	}

	// Service is CRUDL operations interface for the patterns.
	Service interface {

		// Create adds the specified Pattern if not present yet. Returns the created pattern Code, error otherwise.
		// May return ErrConcurrentUpdate, ErrInvalidPattern or ErrInternal when fails.
		Create(ctx context.Context, src string) (model.Pattern, error)

		// Read returns the Pattern by the specified Code if it exists. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Read(ctx context.Context, code model.PatternCode) (model.Pattern, error)

		// Delete removes the pattern if present by the specified Code. Otherwise, returns ErrNotFound.
		// May also return ErrInternal when fails.
		Delete(ctx context.Context, code model.PatternCode) error

		// SearchMatchesBulk finds all patterns Code matching any of the specified metadata values.
		// This is the convenience bulk function returning the codes grouped by the input metadata keys.
		// A client should use the greatest key and last pattern Code from the previous results page to provide a next cursor.
		// Returns the page of results with count not more than the limit specified in the query.
		SearchMatchesBulk(ctx context.Context, md model.Metadata, limit uint32, cursor *BulkCursor) (page map[string][]model.PatternCode, err error)
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

func (svc service) Create(ctx context.Context, src string) (p model.Pattern, err error) {
	req := &apiGrpc.CreateRequest{
		Src: src,
	}
	var resp *apiGrpc.CreateResponse
	resp, err = svc.client.Create(ctx, req)
	if err == nil {
		respPattern := resp.Pattern
		p = model.Pattern{
			Code:  respPattern.Code,
			Regex: respPattern.Regex,
			Src:   respPattern.Src,
		}
	}
	return
}

func (svc service) Read(ctx context.Context, code model.PatternCode) (p model.Pattern, err error) {
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
			p = model.Pattern{
				Code:  respPattern.Code,
				Regex: respPattern.Regex,
				Src:   respPattern.Src,
			}
		}
	}
	return
}

func (svc service) Delete(ctx context.Context, code model.PatternCode) (err error) {
	req := &apiGrpc.DeleteRequest{
		Code: code,
	}
	_, err = svc.client.Delete(ctx, req)
	return
}

func (svc service) SearchMatchesBulk(ctx context.Context, md model.Metadata, limit uint32, cursor *BulkCursor) (page map[string][]model.PatternCode, err error) {
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
		respPage := resp.Page
		page = make(map[string][]model.PatternCode, len(respPage))
		for k, respCodes := range respPage {
			var codes []model.PatternCode
			for _, respCode := range respCodes.Values {
				codes = append(codes, respCode)
			}
			page[k] = codes
		}
	}
	return
}
