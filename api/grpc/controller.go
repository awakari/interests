package grpc

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/api/grpc/common"
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
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		var cond condition.Condition
		cond, err = decodeCondition(req.Cond)
		if err == nil {
			reqMd := req.Md
			md := subscription.Metadata{
				Description: reqMd.Description,
				Enabled:     reqMd.Enabled,
			}
			sd := subscription.Data{
				Metadata:  md,
				Condition: cond,
			}
			resp.Id, err = sc.svc.Create(ctx, groupId, userId, sd)
		}
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *ReadResponse, err error) {
	resp = &ReadResponse{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		var sd subscription.Data
		sd, err = sc.svc.Read(ctx, req.Id, groupId, userId)
		if err == nil {
			resp.Cond = &common.ConditionOutput{}
			encodeCondition(sd.Condition, resp.Cond)
			md := sd.Metadata
			resp.Md = &Metadata{
				Description: md.Description,
				Enabled:     md.Enabled,
			}
		}
		err = encodeError(err)
	}
	return
}

func (sc serviceController) UpdateMetadata(ctx context.Context, req *UpdateMetadataRequest) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		reqMd := req.Md
		md := subscription.Metadata{
			Description: reqMd.Description,
			Enabled:     reqMd.Enabled,
		}
		err = sc.svc.UpdateMetadata(ctx, req.Id, groupId, userId, md)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Delete(ctx context.Context, req *DeleteRequest) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		err = sc.svc.Delete(ctx, req.Id, groupId, userId)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) SearchOwn(ctx context.Context, req *SearchOwnRequest) (resp *SearchOwnResponse, err error) {
	resp = &SearchOwnResponse{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		q := subscription.QueryByAccount{
			GroupId: groupId,
			UserId:  userId,
			Limit:   req.Limit,
		}
		resp.Ids, err = sc.svc.SearchByAccount(ctx, q, req.Cursor)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) SearchByCondition(req *SearchByConditionRequest, server Service_SearchByConditionServer) (err error) {
	var cond condition.Condition
	kcq := req.GetKcq()
	switch {
	case kcq != nil:
		cond = condition.NewKiwiCondition(
			condition.NewKeyCondition(condition.NewCondition(false), "", kcq.Key),
			kcq.Partial,
			kcq.Pattern,
		)
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	if err == nil {
		ctx := server.Context()
		sendToStreamFunc := func(cm *subscription.ConditionMatch) (err error) {
			return sendToStream(cm, server)
		}
		err = sc.svc.SearchByCondition(ctx, cond, sendToStreamFunc)
	}
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
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

func sendToStream(cm *subscription.ConditionMatch, server Service_SearchByConditionServer) (err error) {
	var respCm ConditionMatch
	encodeConditionMatch(cm, &respCm)
	err = server.Send(&respCm)
	return
}

func encodeCondition(src condition.Condition, dst *common.ConditionOutput) {
	dst.Not = src.IsNot()
	switch c := src.(type) {
	case condition.GroupCondition:
		var dstGroup []*common.ConditionOutput
		for _, childSrc := range c.GetGroup() {
			var childDst common.ConditionOutput
			encodeCondition(childSrc, &childDst)
			dstGroup = append(dstGroup, &childDst)
		}
		dst.Cond = &common.ConditionOutput_Gc{
			Gc: &common.GroupConditionOutput{
				Logic: common.GroupLogic(c.GetLogic()),
				Group: dstGroup,
			},
		}
	case condition.KiwiCondition:
		dst.Cond = &common.ConditionOutput_Kc{
			Kc: &common.KiwiConditionOutput{
				Id:      c.GetId(),
				Key:     c.GetKey(),
				Pattern: c.GetPattern(),
				Partial: c.IsPartial(),
			},
		}
	}
	return
}

func encodeConditionMatch(src *subscription.ConditionMatch, dst *ConditionMatch) {
	dst.SubId = src.SubscriptionId
	dst.CondId = src.ConditionId
	dst.Cond = &common.ConditionOutput{}
	encodeCondition(src.Condition, dst.Cond)
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
