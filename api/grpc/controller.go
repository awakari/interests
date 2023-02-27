package grpc

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
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

func (sc serviceController) Create(ctx context.Context, req *SubscriptionDataInput) (resp *CreateResponse, err error) {
	var id string
	route := req.Route
	var cond condition.Condition
	cond, err = decodeCondition(route.Condition)
	if err == nil {
		sd := subscription.Data{
			Metadata: req.Metadata,
			Route: subscription.Route{
				Destinations: route.Destinations,
				Condition:    cond,
			},
		}
		id, err = sc.svc.Create(ctx, sd)
		err = encodeError(err)
	}
	resp = &CreateResponse{
		Id: id,
	}
	return
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *SubscriptionDataOutput, err error) {
	var sd subscription.Data
	sd, err = sc.svc.Read(ctx, req.Id)
	if err != nil {
		err = encodeError(err)
	} else {
		resp, err = encodeSubscriptionData(sd)
	}
	return
}

func (sc serviceController) Delete(ctx context.Context, req *DeleteRequest) (resp *emptypb.Empty, err error) {
	err = sc.svc.Delete(ctx, req.Id)
	return &emptypb.Empty{}, encodeError(err)
}

func (sc serviceController) SearchByCondition(ctx context.Context, req *SearchByConditionRequest) (resp *SearchByConditionResponse, err error) {
	resp = &SearchByConditionResponse{
		Page: []*ConditionMatch{},
	}
	kcq := req.GetKiwiConditionQuery()
	switch {
	case kcq != nil:
		q := condition.Query{
			Limit: req.Limit,
			Condition: condition.NewKiwiCondition(
				condition.NewKeyCondition(condition.NewCondition(false), "", kcq.Key),
				kcq.Partial,
				kcq.Pattern,
			),
		}
		var page []subscription.ConditionMatch
		var encodedCm *ConditionMatch
		page, err = sc.svc.SearchByCondition(ctx, q, req.Cursor)
		for _, cm := range page {
			encodedCm, err = encodeConditionMatch(cm)
			resp.Page = append(resp.Page, encodedCm)
		}
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func (sc serviceController) SearchByMetadata(ctx context.Context, req *SearchByMetadataRequest) (resp *SearchByMetadataResponse, err error) {
	resp = &SearchByMetadataResponse{
		Page: []*Subscription{},
	}
	q := model.MetadataQuery{
		Limit:    req.Limit,
		Metadata: req.Metadata,
	}
	var page []subscription.Subscription
	var encodedSub *Subscription
	page, err = sc.svc.SearchByMetadata(ctx, q, req.Cursor)
	for _, sub := range page {
		encodedSub, err = encodeSubscription(sub)
		resp.Page = append(resp.Page, encodedSub)
	}
	if err != nil {
		err = encodeError(err)
	}
	return
}

func decodeCondition(src *ConditionInput) (dst condition.Condition, err error) {
	gc, ktc := src.GetGroupCondition(), src.GetKiwiTreeCondition()
	switch {
	case gc != nil:
		var group []condition.Condition
		var childDst condition.Condition
		for _, childSrc := range gc.Group {
			childDst, err = decodeCondition(childSrc)
			if err != nil {
				break
			}
			group = append(group, childDst)
		}
		if err == nil {
			dst = condition.NewGroupCondition(
				condition.NewCondition(src.Not),
				condition.GroupLogic(gc.GetLogic()),
				group,
			)
		}
	case ktc != nil:
		dst = condition.NewKiwiTreeCondition(
			condition.NewKiwiCondition(
				condition.NewKeyCondition(condition.NewCondition(src.Not), "", ktc.GetKey()),
				ktc.GetPartial(),
				ktc.GetPattern(),
			),
		)
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func encodeSubscription(src subscription.Subscription) (dst *Subscription, err error) {
	var dstData *SubscriptionDataOutput
	dstData, err = encodeSubscriptionData(src.Data)
	if err == nil {
		dst = &Subscription{
			Id:   src.Id,
			Data: dstData,
		}
	}
	return
}

func encodeSubscriptionData(src subscription.Data) (dst *SubscriptionDataOutput, err error) {
	var dstRoute *RouteOutput
	dstRoute, err = encodeSubscriptionRoute(src.Route)
	if err == nil {
		dst = &SubscriptionDataOutput{
			Metadata: src.Metadata,
			Route:    dstRoute,
		}
	}
	return
}

func encodeSubscriptionRoute(src subscription.Route) (dst *RouteOutput, err error) {
	var dstCond *ConditionOutput
	dstCond, err = encodeCondition(src.Condition)
	if err == nil {
		dst = &RouteOutput{
			Destinations: src.Destinations,
			Condition:    dstCond,
		}
	}
	return
}

func encodeCondition(src condition.Condition) (dst *ConditionOutput, err error) {
	dst = &ConditionOutput{
		Not: src.IsNot(),
	}
	switch c := src.(type) {
	case condition.GroupCondition:
		var dstGroup []*ConditionOutput
		var childDst *ConditionOutput
		for _, childSrc := range c.GetGroup() {
			childDst, err = encodeCondition(childSrc)
			if err != nil {
				break
			}
			dstGroup = append(dstGroup, childDst)
		}
		if err == nil {
			dst.Condition = &ConditionOutput_GroupCondition{
				GroupCondition: &GroupConditionOutput{
					Logic: GroupLogic(c.GetLogic()),
					Group: dstGroup,
				},
			}
		}
	case condition.KiwiCondition:
		dst.Condition = &ConditionOutput_KiwiCondition{
			KiwiCondition: &KiwiConditionOutput{
				Id:      c.GetId(),
				Key:     c.GetKey(),
				Pattern: c.GetPattern(),
				Partial: c.IsPartial(),
			},
		}
	}
	return
}

func encodeConditionMatch(src subscription.ConditionMatch) (dst *ConditionMatch, err error) {
	var dstRoute *RouteOutput
	dstRoute, err = encodeSubscriptionRoute(src.Route)
	dst = &ConditionMatch{
		SubscriptionId: src.Id,
		ConditionId:    src.ConditionId,
		Route:          dstRoute,
	}
	return
}

func encodeError(svcErr error) (err error) {
	switch {
	case svcErr == nil:
		err = nil
	case errors.Is(svcErr, service.ErrInternal):
		err = status.Error(codes.Internal, svcErr.Error())
	case errors.Is(svcErr, subscription.ErrInvalidSubscriptionRoute):
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
