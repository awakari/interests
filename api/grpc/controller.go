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

func (sc serviceController) Create(ctx context.Context, req *SubscriptionInputData) (resp *CreateResponse, err error) {
	var id string
	var cond model.Condition
	cond, err = decodeCondition(req.Condition)
	if err == nil {
		sd := model.SubscriptionData{
			Metadata:  req.Metadata,
			Routes:    req.Routes,
			Condition: cond,
		}
		id, err = sc.svc.Create(ctx, sd)
		err = encodeError(err)
	}
	resp = &CreateResponse{
		Id: id,
	}
	return
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *SubscriptionOutputData, err error) {
	var sd model.SubscriptionData
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

func (sc serviceController) SearchByCondition(ctx context.Context, req *SearchByConditionRequest) (resp *SearchResponse, err error) {
	resp = &SearchResponse{
		Page: []*Subscription{},
	}
	kcq := req.GetKiwiConditionQuery()
	switch {
	case kcq != nil:
		q := model.ConditionQuery{
			Limit: req.Limit,
			Condition: model.NewKiwiCondition(
				model.NewKeyCondition(
					model.NewCondition(false), // not flag is not used in the search by condition
					kcq.Key,
				),
				kcq.Partial,
				kcq.Pattern,
			),
		}
		var page []model.Subscription
		var encodedSub *Subscription
		page, err = sc.svc.SearchByCondition(ctx, q, req.Cursor)
		for _, sub := range page {
			encodedSub, err = encodeSubscription(sub)
			resp.Page = append(resp.Page, encodedSub)
		}
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func (sc serviceController) SearchByMetadata(ctx context.Context, req *SearchByMetadataRequest) (resp *SearchResponse, err error) {
	resp = &SearchResponse{
		Page: []*Subscription{},
	}
	q := model.MetadataQuery{
		Limit:    req.Limit,
		Metadata: req.Metadata,
	}
	var page []model.Subscription
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
			dst = model.NewGroupCondition(
				model.NewCondition(gc.GetNot()),
				model.GroupLogic(gc.GetLogic()),
				group,
			)
		}
	case ktc != nil:
		dst = model.NewKiwiTreeCondition(
			model.NewKiwiCondition(
				model.NewKeyCondition(
					model.NewCondition(ktc.GetNot()),
					ktc.GetKey(),
				),
				ktc.GetPartial(),
				ktc.GetPattern(),
			),
		)
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func encodeSubscription(src model.Subscription) (dst *Subscription, err error) {
	var dstData *SubscriptionOutputData
	dstData, err = encodeSubscriptionData(src.Data)
	if err == nil {
		dst = &Subscription{
			Id:   src.Id,
			Data: dstData,
		}
	}
	return
}

func encodeSubscriptionData(src model.SubscriptionData) (dst *SubscriptionOutputData, err error) {
	var dstCond *OutputCondition
	dstCond, err = encodeCondition(src.Condition)
	dst = &SubscriptionOutputData{
		Metadata:  src.Metadata,
		Routes:    src.Routes,
		Condition: dstCond,
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
					Base: &OutputConditionBase{
						Not: src.IsNot(),
					},
					Logic: GroupLogic(c.GetLogic()),
					Group: dstGroup,
				},
			}
		}
	case model.KiwiCondition:
		dst.Condition = &OutputCondition_KiwiCondition{
			KiwiCondition: &KiwiOutputCondition{
				Base: &OutputConditionBase{
					Not: c.IsNot(),
				},
				Key:     c.GetKey(),
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
