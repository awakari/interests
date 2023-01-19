package grpc

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/service"
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
	resp = &emptypb.Empty{}
	var cond model.Condition
	cond, err = decodeCondition(req.Condition)
	if err == nil {
		sub := model.Subscription{
			Name:        req.Name,
			Description: req.Description,
			Routes:      req.Routes,
			Condition:   cond,
		}
		err = sc.svc.Create(ctx, sub)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *Subscription, err error) {
	var sub model.Subscription
	sub, err = sc.svc.Read(ctx, req.Name)
	if err != nil {
		err = encodeError(err)
	} else {
		resp, err = encodeSubscription(sub)
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

func (sc serviceController) SearchByKiwi(ctx context.Context, req *SearchByKiwiRequest) (resp *SearchResponse, err error) {
	var page []model.Subscription
	q := model.KiwiQuery{
		Limit:   req.Limit,
		Key:     req.Key,
		Pattern: req.Pattern,
		Partial: req.Partial,
	}
	page, err = sc.svc.SearchByKiwi(ctx, q, req.Cursor)
	if err != nil {
		err = encodeError(err)
	} else {
		var respSubs []*Subscription
		var respSub *Subscription
		for _, sub := range page {
			respSub, err = encodeSubscription(sub)
			if err != nil {
				break
			}
			respSubs = append(respSubs, respSub)
		}
		resp = &SearchResponse{
			Page: respSubs,
		}
	}
	return
}

func decodeCondition(src *InputCondition) (dst model.Condition, err error) {
	gc, ktc := src.GetGroupCondition(), src.GetKiwiTreeCondition()
	switch {
	case gc != nil:
		var group []model.Condition
		var childDst model.Condition
		for _, childSrc := range gc.Group {
			childDst, err = decodeCondition(childSrc)
			if err != nil {
				break
			}
			group = append(group, childDst)
		}
		if err == nil {
			gcBase := gc.GetBase()
			dst = model.NewGroupCondition(
				model.NewCondition(gcBase.GetBase().GetNot()),
				model.GroupLogic(gcBase.GetLogic()),
				group,
			)
		}
	case ktc != nil:
		dst = model.NewKiwiTreeCondition(
			model.NewKiwiCondition(
				model.NewKeyCondition(
					model.NewCondition(ktc.GetBase().GetBase().GetBase().GetNot()),
					ktc.GetBase().GetBase().GetKey(),
				),
				ktc.GetBase().GetPartial(),
				ktc.GetBase().GetPattern(),
			),
		)
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func encodeSubscription(src model.Subscription) (dst *Subscription, err error) {
	var dstCond *OutputCondition
	dstCond, err = encodeCondition(src.Condition)
	dst = &Subscription{
		Name:        src.Name,
		Description: src.Description,
		Routes:      src.Routes,
		Condition:   dstCond,
	}
	return
}

func encodeCondition(src model.Condition) (dst *OutputCondition, err error) {
	dst = &OutputCondition{}
	switch c := src.(type) {
	case model.GroupCondition:
		var dstGroup []*OutputCondition
		var childDst *OutputCondition
		for _, childSrc := range c.GetGroup() {
			childDst, err = encodeCondition(childSrc)
			if err != nil {
				break
			}
			dstGroup = append(dstGroup, childDst)
		}
		if err == nil {
			dst.Condition = &OutputCondition_GroupCondition{
				GroupCondition: &GroupOutputCondition{
					Base: &GroupConditionBase{
						Base: &ConditionBase{
							Not: src.IsNot(),
						},
						Logic: GroupLogic(c.GetLogic()),
					},
					Group: dstGroup,
				},
			}
		}
	case model.KiwiCondition:
		dst.Condition = &OutputCondition_KiwiCondition{
			KiwiCondition: &KiwiCondition{
				Base: &KeyCondition{
					Base: &ConditionBase{
						Not: c.IsNot(),
					},
					Key: c.GetKey(),
				},
				Pattern: c.GetPattern(),
				Partial: c.IsPartial(),
			},
		}
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
