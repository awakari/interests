package grpc

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serviceController struct {
	svc service.Service
}

func NewServiceController(svc service.Service) ServiceServer {
	return serviceController{
		svc: svc,
	}
}

func (sc serviceController) Create(ctx context.Context, req *CreateRequest) (resp *CreateResponse, err error) {
	resp = &CreateResponse{}
	var cond condition.Condition
	cond, err = decodeCondition(req.Cond)
	if err == nil {
		reqMd := req.Md
		md := subscription.Metadata{
			Description: reqMd.Description,
			Priority:    reqMd.Priority,
			Enabled:     reqMd.Enabled,
		}
		sd := subscription.Data{
			Metadata:  md,
			Condition: cond,
		}
		resp.Id, err = sc.svc.Create(ctx, ctx.Value(keyAccount).(string), sd)
	}
	err = encodeError(err)
	return
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *ReadResponse, err error) {
	resp = &ReadResponse{}
	var sd subscription.Data
	sd, err = sc.svc.Read(ctx, req.Id, ctx.Value(keyAccount).(string))
	if err == nil {
		resp.Cond = encodeCondition(sd.Condition)
		md := sd.Metadata
		resp.Md = &Metadata{
			Description: md.Description,
			Priority:    md.Priority,
			Enabled:     md.Enabled,
		}
	}
	if err != nil {
		err = encodeError(err)
	}
	return
}

func (sc serviceController) UpdateMetadata(ctx context.Context, req *UpdateMetadataRequest) (resp *emptypb.Empty, err error) {
	reqMd := req.Md
	md := subscription.Metadata{
		Description: reqMd.Description,
		Priority:    reqMd.Priority,
		Enabled:     reqMd.Enabled,
	}
	err = sc.svc.UpdateMetadata(ctx, req.Id, ctx.Value(keyAccount).(string), md)
	return &emptypb.Empty{}, encodeError(err)
}

func (sc serviceController) Delete(ctx context.Context, req *DeleteRequest) (resp *emptypb.Empty, err error) {
	err = sc.svc.Delete(ctx, req.Id, ctx.Value(keyAccount).(string))
	return &emptypb.Empty{}, encodeError(err)
}

func (sc serviceController) SearchOwn(ctx context.Context, req *SearchOwnRequest) (resp *SearchOwnResponse, err error) {
	q := subscription.QueryByAccount{
		Account: ctx.Value(keyAccount).(string),
		Limit:   req.Limit,
	}
	resp = &SearchOwnResponse{}
	resp.Ids, err = sc.svc.SearchByAccount(ctx, q, req.Cursor)
	if err != nil {
		err = encodeError(err)
	}
	return
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
		var respCm *ConditionMatch
		page, err = sc.svc.SearchByCondition(ctx, q, req.Cursor)
		for _, cm := range page {
			respCm = encodeConditionMatch(cm)
			resp.Page = append(resp.Page, respCm)
		}
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func decodeCondition(src *ConditionInput) (dst condition.Condition, err error) {
	gc, ktc := src.GetGc(), src.GetKtc()
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

func encodeCondition(src condition.Condition) (dst *ConditionOutput) {
	dst = &ConditionOutput{
		Not: src.IsNot(),
	}
	switch c := src.(type) {
	case condition.GroupCondition:
		var dstGroup []*ConditionOutput
		var childDst *ConditionOutput
		for _, childSrc := range c.GetGroup() {
			childDst = encodeCondition(childSrc)
			dstGroup = append(dstGroup, childDst)
		}
		dst.Condition = &ConditionOutput_GroupCondition{
			GroupCondition: &GroupConditionOutput{
				Logic: GroupLogic(c.GetLogic()),
				Group: dstGroup,
			},
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

func encodeConditionMatch(src subscription.ConditionMatch) *ConditionMatch {
	return &ConditionMatch{
		SubscriptionId: src.Id,
		ConditionId:    src.ConditionId,
		Condition:      encodeCondition(src.Condition),
	}
}

func encodeError(svcErr error) (err error) {
	switch {
	case svcErr == nil:
		err = nil
	case errors.Is(svcErr, service.ErrInternal):
		err = status.Error(codes.Internal, svcErr.Error())
	case errors.Is(svcErr, subscription.ErrInvalidSubscriptionCondition):
		err = status.Error(codes.InvalidArgument, svcErr.Error())
	case errors.Is(svcErr, service.ErrShouldRetry):
		err = status.Error(codes.Unavailable, svcErr.Error())
	case errors.Is(svcErr, service.ErrNotFound):
		err = status.Error(codes.NotFound, svcErr.Error())
	default:
		err = status.Error(codes.Internal, svcErr.Error())
	}
	return
}
