package matchers

import (
	"context"
	"errors"
	"fmt"
	grpcApi "github.com/meandros-messaging/subscriptions/api/grpc/matchers"
	"github.com/meandros-messaging/subscriptions/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (

	// Service represents the matchers service.
	Service interface {

		// Create creates a new Pattern and registers the corresponding model.MatcherData.
		// It increments the model.MatcherData reference count by 1 if it exists already.
		// Repeating the same Create call yields the same state.
		Create(ctx context.Context, k string, patternSrc string) (m model.MatcherData, err error)

		// Delete decrements the model.MatcherData reference count by 1.
		// It removes a specified model.MatcherData from the underlying storage if the resulting reference count is less
		// than 1.
		Delete(ctx context.Context, m model.MatcherData) (err error)

		// Search returns all matchers those match the specified key/value pair.
		// Results are paginated, use the last pattern code from the previous page as a cursor to get the next.
		Search(ctx context.Context, k, v string, limit uint32, cursor model.PatternCode) (page []model.PatternCode, err error)
	}

	service struct {
		client grpcApi.ServiceClient
	}
)

var (

	// ErrInvalidPatternSrc indicates the source string to create a Pattern is invalid.
	ErrInvalidPatternSrc = errors.New("invalid pattern source")

	// ErrShouldRetry indicates a storage entity is locked and the operation should be retried.
	ErrShouldRetry = errors.New("retry the operation")

	// ErrInternal indicates some unexpected internal failure.
	ErrInternal = errors.New("internal failure")

	// ErrNotFound indicates the specified Matcher was not found in the underlying storage.
	ErrNotFound = errors.New("not found")
)

func NewService(client grpcApi.ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Create(ctx context.Context, k string, patternSrc string) (m model.MatcherData, err error) {
	req := &grpcApi.CreateRequest{
		Key:        k,
		PatternSrc: patternSrc,
	}
	var resp *grpcApi.CreateResponse
	resp, err = svc.client.Create(ctx, req)
	if err != nil {
		err = decodeError(err)
	} else {
		p := model.Pattern{
			Code: resp.Matcher.PatternCode,
			Src:  patternSrc,
		}
		m = model.MatcherData{
			Key:     resp.Matcher.Key,
			Pattern: p,
		}
	}
	return
}

func (svc service) Delete(ctx context.Context, m model.MatcherData) (err error) {
	req := &grpcApi.DeleteRequest{
		Matcher: &grpcApi.MatcherData{
			Key:         m.Key,
			PatternCode: m.Pattern.Code,
		},
	}
	_, err = svc.client.Delete(ctx, req)
	if err != nil {
		err = decodeError(err)
	}
	return err
}

func (svc service) Search(ctx context.Context, k, v string, limit uint32, cursor model.PatternCode) (page []model.PatternCode, err error) {
	req := &grpcApi.SearchRequest{
		Query: &grpcApi.Query{
			Key:   k,
			Value: v,
			Limit: limit,
		},
		Cursor: cursor,
	}
	var resp *grpcApi.SearchResponse
	resp, err = svc.client.Search(ctx, req)
	if err != nil {
		err = decodeError(err)
	} else {
		for _, pc := range resp.Page {
			page = append(page, pc)
		}
	}
	return
}

func decodeError(grpcErr error) (svcErr error) {
	switch status.Code(grpcErr) {
	case codes.InvalidArgument:
		svcErr = fmt.Errorf("%w: %s", ErrInvalidPatternSrc, grpcErr)
	case codes.NotFound:
		svcErr = fmt.Errorf("%w: %s", ErrNotFound, grpcErr)
	case codes.Internal:
		svcErr = fmt.Errorf("%w: %s", ErrInternal, grpcErr)
	case codes.Unavailable:
		svcErr = fmt.Errorf("%w: %s", ErrShouldRetry, grpcErr)
	default:
		svcErr = fmt.Errorf("%w: %s", ErrInternal, grpcErr)
	}
	return
}
