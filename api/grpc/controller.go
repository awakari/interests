package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/interests/api/grpc/common"
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/model/interest"
	"github.com/awakari/interests/storage"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
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
			sd := interest.Data{
				Description: req.Description,
				Enabled:     req.Enabled,
				Condition:   cond,
				Created:     time.Now().UTC(),
				Public:      req.Public,
			}
			// check is for the backward compatibility
			if req.Expires != nil {
				sd.Expires = req.Expires.AsTime()
			}
			switch req.Id {
			case "":
				resp.Id = ksuid.New().String()
			default:
				resp.Id = req.Id
			}
			err = sc.stor.Create(ctx, resp.Id, groupId, userId, sd)
		}
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Read(ctx context.Context, req *ReadRequest) (resp *ReadResponse, err error) {
	resp = &ReadResponse{}
	var groupId string
	var userId string
	if !req.Internal {
		groupId, userId, err = getAuthInfo(ctx)
	}
	if err == nil {
		var sd interest.Data
		var ownerGroupId string
		var ownerUserId string
		sd, ownerGroupId, ownerUserId, err = sc.stor.Read(ctx, req.Id, groupId, userId, req.Internal)
		if err == nil {
			resp.Cond = &Condition{}
			encodeCondition(sd.Condition, resp.Cond)
			resp.Description = sd.Description
			resp.Enabled = sd.Enabled
			if !sd.EnabledSince.IsZero() {
				resp.EnabledSince = timestamppb.New(sd.EnabledSince)
			}
			resp.Public = sd.Public
			resp.Followers = sd.Followers
			if !sd.Expires.IsZero() {
				resp.Expires = timestamppb.New(sd.Expires)
			}
			if !sd.Created.IsZero() {
				resp.Created = timestamppb.New(sd.Created)
			}
			if !sd.Updated.IsZero() {
				resp.Updated = timestamppb.New(sd.Updated)
			}
			if !sd.Result.IsZero() {
				resp.Result = timestamppb.New(sd.Result)
			}
			resp.GroupId = ownerGroupId
			resp.UserId = ownerUserId
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
	var cond condition.Condition
	if err == nil {
		cond, err = decodeCondition(req.Cond)
	}
	if err == nil {
		sd := interest.Data{
			Description: req.Description,
			Enabled:     req.Enabled,
			Condition:   cond,
			Updated:     time.Now().UTC(),
			Public:      req.Public,
		}
		// check is for the backward compatibility
		if req.Expires != nil {
			sd.Expires = req.Expires.AsTime()
		}
		var prev interest.Data
		prev, err = sc.stor.Update(ctx, req.Id, groupId, userId, sd)
		if err == nil {
			resp.Cond = &Condition{}
			encodeCondition(prev.Condition, resp.Cond)
		}
		err = encodeError(err)
	}
	return
}

func (sc serviceController) UpdateFollowers(ctx context.Context, req *UpdateFollowersRequest) (resp *UpdateFollowersResponse, err error) {
	resp = &UpdateFollowersResponse{}
	err = sc.stor.UpdateFollowers(ctx, req.Id, req.Count)
	err = encodeError(err)
	return
}

func (sc serviceController) UpdateResultTime(ctx context.Context, req *UpdateResultTimeRequest) (resp *UpdateResultTimeResponse, err error) {
	resp = &UpdateResultTimeResponse{}
	switch req.Read {
	case nil:
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("interest %s update result time missing argument", req.Id))
	default:
		err = sc.stor.UpdateResultTime(ctx, req.Id, req.Read.AsTime().UTC())
		err = encodeError(err)
	}
	return
}

func (sc serviceController) SetEnabledBatch(ctx context.Context, req *SetEnabledBatchRequest) (resp *SetEnabledBatchResponse, err error) {
	resp = &SetEnabledBatchResponse{}
	var enabledSince time.Time
	if req.EnabledSince != nil && req.EnabledSince.IsValid() {
		enabledSince = req.EnabledSince.AsTime().UTC()
	}
	resp.N, err = sc.stor.SetEnabledBatch(ctx, req.Ids, req.Enabled, enabledSince)
	err = encodeError(err)
	return
}

func (sc serviceController) Delete(ctx context.Context, req *DeleteRequest) (resp *DeleteResponse, err error) {
	resp = &DeleteResponse{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		var sd interest.Data
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
		q := interest.Query{
			GroupId:     groupId,
			UserId:      userId,
			Limit:       req.Limit,
			Pattern:     req.Pattern,
			PrivateOnly: req.Private,
		}
		switch req.Order {
		case Order_DESC:
			q.Order = interest.OrderDesc
		default:
			q.Order = interest.OrderAsc
		}
		resp.Ids, err = sc.stor.Search(ctx, q, interest.Cursor{
			Id: req.Cursor,
		})
		err = encodeError(err)
	}
	return
}

func (sc serviceController) Search(ctx context.Context, req *SearchRequest) (resp *SearchResponse, err error) {
	resp = &SearchResponse{}
	var groupId string
	var userId string
	groupId, userId, err = getAuthInfo(ctx)
	if err == nil {
		q := interest.Query{
			GroupId:       groupId,
			UserId:        userId,
			Limit:         req.Limit,
			Pattern:       req.Pattern,
			IncludePublic: true,
		}
		switch req.Sort {
		case Sort_FOLLOWERS:
			q.Sort = interest.SortFollowers
		case Sort_TIME_CREATED:
			q.Sort = interest.SortTimeCreated
		default:
			q.Sort = interest.SortId
		}
		switch req.Order {
		case Order_DESC:
			q.Order = interest.OrderDesc
		default:
			q.Order = interest.OrderAsc
		}
		var cursor interest.Cursor
		if req.Cursor != nil {
			cursor.Id = req.Cursor.Id
			cursor.Followers = req.Cursor.Followers
			if req.Cursor.TimeCreated != nil {
				cursor.CreatedAt = req.Cursor.TimeCreated.AsTime().UTC()
			}
		}
		resp.Ids, err = sc.stor.Search(ctx, q, cursor)
		err = encodeError(err)
	}
	return
}

func (sc serviceController) SearchByCondition(ctx context.Context, req *SearchByConditionRequest) (resp *SearchByConditionResponse, err error) {
	q := interest.QueryByCondition{
		CondId: req.CondId,
		Limit:  req.Limit,
	}
	var page interest.ConditionMatchPage
	page, err = sc.stor.SearchByCondition(ctx, q, req.Cursor)
	if err == nil {
		resp = &SearchByConditionResponse{
			Expires: timestamppb.New(page.Expires),
		}
		for _, cm := range page.ConditionMatches {
			result := SearchByConditionResult{}
			encodeConditionMatch(cm, &result)
			resp.Page = append(resp.Page, &result)
		}
	}
	err = encodeError(err)
	return
}

func decodeCondition(src *Condition) (dst condition.Condition, err error) {
	gc, tc, nc := src.GetGc(), src.GetTc(), src.GetNc()
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
	case nc != nil:
		dst = condition.NewNumberCondition(
			condition.NewKeyCondition(condition.NewCondition(src.Not), nc.Id, nc.Key),
			decodeNumOp(nc.Op),
			nc.Val,
		)
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func decodeNumOp(src Operation) (dst condition.NumOp) {
	switch src {
	case Operation_Gt:
		dst = condition.NumOpGt
	case Operation_Gte:
		dst = condition.NumOpGte
	case Operation_Eq:
		dst = condition.NumOpEq
	case Operation_Lte:
		dst = condition.NumOpLte
	case Operation_Lt:
		dst = condition.NumOpLt
	default:
		dst = condition.NumOpUndefined
	}
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
	case condition.NumberCondition:
		dst.Cond = &Condition_Nc{
			Nc: &NumberCondition{
				Id:  c.GetId(),
				Key: c.GetKey(),
				Op:  encodeNumOp(c.GetOperation()),
				Val: c.GetValue(),
			},
		}
	}
	return
}

func encodeNumOp(src condition.NumOp) (dst Operation) {
	switch src {
	case condition.NumOpGt:
		dst = Operation_Gt
	case condition.NumOpGte:
		dst = Operation_Gte
	case condition.NumOpEq:
		dst = Operation_Eq
	case condition.NumOpLte:
		dst = Operation_Lte
	case condition.NumOpLt:
		dst = Operation_Lt
	default:
		dst = Operation_Undefined
	}
	return
}

func encodeConditionMatch(src interest.ConditionMatch, dst *SearchByConditionResult) {
	dst.Id = src.InterestId
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
	case errors.Is(svcErr, storage.ErrConflict):
		err = status.Error(codes.AlreadyExists, svcErr.Error())
	default:
		err = status.Error(codes.Internal, svcErr.Error())
	}
	return
}
