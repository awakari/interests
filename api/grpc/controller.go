package grpc

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"github.com/meandros-messaging/subscriptions/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	serviceController struct {
		svc service.Service
	}
)

func NewServiceController(svc service.Service) ServiceServer {
	return serviceController{
		svc: svc,
	}
}

func (sc serviceController) Create(ctx context.Context, req *CreateRequest) (resp *emptypb.Empty, err error) {
	var excludes model.MatcherGroup
	if req.Excludes != nil {
		excludes = model.MatcherGroup{
			All:      req.Excludes.All,
			Matchers: decodeMatchers(req.Excludes.Matchers),
		}
	}
	var includes model.MatcherGroup
	if req.Includes != nil {
		includes = model.MatcherGroup{
			All:      req.Includes.All,
			Matchers: decodeMatchers(req.Includes.Matchers),
		}
	}
	createReq := service.CreateRequest{
		Description: req.Description,
		Routes:      req.Routes,
		Includes:    includes,
		Excludes:    excludes,
	}
	err = sc.svc.Create(ctx, req.Name, createReq)
	return &emptypb.Empty{}, encodeError(err)
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *Subscription, err error) {
	var sub model.Subscription
	sub, err = sc.svc.Read(ctx, req.Name)
	if err != nil {
		err = encodeError(err)
	} else {
		resp = encodeSubscription(sub)
	}
	return
}

func (sc serviceController) Delete(ctx context.Context, req *DeleteRequest) (resp *emptypb.Empty, err error) {
	err = sc.svc.Delete(ctx, req.Name)
	return &emptypb.Empty{}, encodeError(err)
}

func (sc serviceController) ListNames(ctx context.Context, req *ListNamesRequest) (resp *ListNamesResponse, err error) {
	var page []string
	page, err = sc.svc.ListNames(ctx, req.Limit, req.Cursor)
	if err != nil {
		err = encodeError(err)
	} else {
		resp = &ListNamesResponse{
			Names: page,
		}
	}
	return
}

func (sc serviceController) Search(ctx context.Context, req *SearchRequest) (resp *SearchResponse, err error) {
	var page []model.Subscription
	m := decodeMatcher(req.Matcher)
	q := service.Query{
		Limit:      req.Limit,
		InExcludes: req.InExcludes,
		Matcher:    m,
	}
	page, err = sc.svc.Search(ctx, q, req.Cursor)
	if err != nil {
		err = encodeError(err)
	} else {
		respSubs := encodeSubscriptions(page)
		resp = &SearchResponse{
			Page: respSubs,
		}
	}
	return
}

func decodeMatchers(reqMatchers []*Matcher) (matchers []model.Matcher) {
	for _, reqMatcher := range reqMatchers {
		m := decodeMatcher(reqMatcher)
		matchers = append(matchers, m)
	}
	return
}

func decodeMatcher(reqMatcher *Matcher) (m model.Matcher) {
	if reqMatcher != nil {
		var p model.Pattern
		if reqMatcher.Pattern != nil {
			p.Code = reqMatcher.Pattern.Code
			p.Src = reqMatcher.Pattern.Src
		}
		m.Partial = reqMatcher.Partial
		m.Key = reqMatcher.Key
		m.Pattern = p
	}
	return
}

func encodeSubscriptions(subs []model.Subscription) (resp []*Subscription) {
	for _, sub := range subs {
		respSub := encodeSubscription(sub)
		resp = append(resp, respSub)
	}
	return
}

func encodeSubscription(sub model.Subscription) (resp *Subscription) {
	resp = &Subscription{
		Name:        sub.Name,
		Description: sub.Description,
		Routes:      sub.Routes,
		Excludes:    encodeMatcherGroup(sub.Excludes),
		Includes:    encodeMatcherGroup(sub.Includes),
	}
	return
}

func encodeMatcherGroup(mg model.MatcherGroup) (resp *MatcherGroup) {
	resp = &MatcherGroup{
		All:      mg.All,
		Matchers: encodeMatchers(mg.Matchers),
	}
	return
}

func encodeMatchers(matchers []model.Matcher) (resp []*Matcher) {
	for _, m := range matchers {
		respMatcher := &Matcher{
			Partial: m.Partial,
			Key:     m.Key,
			Pattern: &Pattern{
				Code: m.Pattern.Code,
				Src:  m.Pattern.Src,
			},
		}
		resp = append(resp, respMatcher)
	}
	return
}

func encodeError(svcErr error) (err error) {
	switch {
	case svcErr == nil:
		err = nil
	case errors.Is(svcErr, service.ErrInternal):
		err = status.Error(codes.Internal, svcErr.Error())
	case errors.Is(svcErr, model.ErrInvalidSubscription):
		err = status.Error(codes.InvalidArgument, svcErr.Error())
	case errors.Is(svcErr, service.ErrShouldRetry):
		err = status.Error(codes.Unavailable, svcErr.Error())
	case errors.Is(svcErr, service.ErrNotFound):
		err = status.Error(codes.NotFound, svcErr.Error())
	case errors.Is(svcErr, service.ErrConflict):
		err = status.Error(codes.AlreadyExists, svcErr.Error())
	default:
		err = status.Error(codes.Internal, svcErr.Error())
	}
	return
}
