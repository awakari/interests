package grpc

import (
	"context"
	"errors"
	"github.com/awakari/subscriptions/api/grpc/common"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serviceController struct {
	stor storage.Storage
}

func NewServiceController(stor storage.Storage) ServiceServer {
	return serviceController{
		stor: stor,
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
			sd := subscription.Data{
				Description: req.Description,
				Enabled:     req.Enabled,
				Condition:   cond,
			}
			resp.Id, err = sc.stor.Create(ctx, groupId, userId, sd)
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
		sd, err = sc.stor.Read(ctx, req.Id, groupId, userId)
		if err == nil {
			resp.Cond = &Condition{}
			encodeCondition(sd.Condition, resp.Cond)
			resp.Description = sd.Description
			resp.Enabled = sd.Enabled
		}
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Update(ctx context.Context, req *UpdateRequest) (resp *UpdateResponse, err error) {
	resp = &UpdateResponse{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		sd := subscription.Data{
			Description: req.Description,
			Enabled:     req.Enabled,
		}
		err = sc.stor.Update(ctx, req.Id, groupId, userId, sd)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Delete(ctx context.Context, req *DeleteRequest) (resp *DeleteResponse, err error) {
	resp = &DeleteResponse{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		var sd subscription.Data
		sd, err = sc.stor.Delete(ctx, req.Id, groupId, userId)
		if err == nil {
			resp.Cond = &Condition{}
			encodeCondition(sd.Condition, resp.Cond)
		}
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
		q := subscription.QueryOwn{
			GroupId: groupId,
			UserId:  userId,
			Limit:   req.Limit,
		}
		resp.Ids, err = sc.stor.SearchOwn(ctx, q, req.Cursor)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) SearchByCondition(req *SearchByConditionRequest, stream Service_SearchByConditionServer) (err error) {
	ctx := stream.Context()
	sendToStreamFunc := func(cm *subscription.ConditionMatch) (err error) {
		return sendToStream(cm, stream)
	}
	err = sc.stor.SearchByCondition(ctx, req.CondId, sendToStreamFunc)
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
	}
	return
}

func decodeCondition(src *Condition) (dst condition.Condition, err error) {
	gc, tc := src.GetGc(), src.GetTc()
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
	case tc != nil:
		dst = condition.NewTextCondition(
			condition.NewKeyCondition(condition.NewCondition(src.Not), tc.GetId(), tc.GetKey()),
			tc.GetTerm(),
			tc.GetExact(),
		)
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func sendToStream(cm *subscription.ConditionMatch, server Service_SearchByConditionServer) (err error) {
	var resp SearchByConditionResponse
	encodeConditionMatch(cm, &resp)
	err = server.Send(&resp)
	return
}

func encodeCondition(src condition.Condition, dst *Condition) {
	dst.Not = src.IsNot()
	switch c := src.(type) {
	case condition.GroupCondition:
		var dstGroup []*Condition
		for _, childSrc := range c.GetGroup() {
			var childDst Condition
			encodeCondition(childSrc, &childDst)
			dstGroup = append(dstGroup, &childDst)
		}
		dst.Cond = &Condition_Gc{
			Gc: &GroupCondition{
				Logic: common.GroupLogic(c.GetLogic()),
				Group: dstGroup,
			},
		}
	case condition.TextCondition:
		dst.Cond = &Condition_Tc{
			Tc: &TextCondition{
				Id:    c.GetId(),
				Key:   c.GetKey(),
				Term:  c.GetTerm(),
				Exact: c.IsExact(),
			},
		}
	}
	return
}

func encodeConditionMatch(src *subscription.ConditionMatch, dst *SearchByConditionResponse) {
	dst.Id = src.SubscriptionId
	dst.Cond = &Condition{}
	encodeCondition(src.Condition, dst.Cond)
}

func encodeError(svcErr error) (err error) {
	switch {
	case svcErr == nil:
		err = nil
	case errors.Is(svcErr, storage.ErrInternal):
		err = status.Error(codes.Internal, svcErr.Error())
	case errors.Is(svcErr, storage.ErrNotFound):
		err = status.Error(codes.NotFound, svcErr.Error())
	default:
		err = status.Error(codes.Internal, svcErr.Error())
	}
	return
}
