package private

import (
	"context"
	"github.com/awakari/subscriptions/api/grpc/common"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serviceController struct {
	svc service.Service
}

func NewServiceController(svc service.Service) ServiceServer {
	return serviceController{
		svc: svc,
	}
}

func (sc serviceController) SearchByCondition(ctx context.Context, req *SearchByConditionRequest) (resp *SearchByConditionResponse, err error) {
	resp = &SearchByConditionResponse{
		Page: []*ConditionMatch{},
	}
	kcq := req.GetKcq()
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
		var cursor subscription.ConditionMatchKey
		reqCursor := req.Cursor
		if reqCursor != nil {
			cursor.Id = reqCursor.SubId
			cursor.Priority = reqCursor.Priority
		}
		page, err = sc.svc.SearchByCondition(ctx, q, cursor)
		if err != nil {
			err = status.Error(codes.Internal, err.Error())
		} else {
			for _, cm := range page {
				respCm = encodeConditionMatch(cm)
				resp.Page = append(resp.Page, respCm)
			}
		}
	default:
		err = status.Error(codes.InvalidArgument, "unsupported condition type")
	}
	return
}

func encodeCondition(src condition.Condition) (dst *common.ConditionOutput) {
	dst = &common.ConditionOutput{
		Not: src.IsNot(),
	}
	switch c := src.(type) {
	case condition.GroupCondition:
		var dstGroup []*common.ConditionOutput
		var childDst *common.ConditionOutput
		for _, childSrc := range c.GetGroup() {
			childDst = encodeCondition(childSrc)
			dstGroup = append(dstGroup, childDst)
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

func encodeConditionMatch(src subscription.ConditionMatch) (dst *ConditionMatch) {
	srcKey := src.Key
	dstKey := &ConditionMatchKey{
		SubId:    srcKey.Id,
		Priority: srcKey.Priority,
	}
	return &ConditionMatch{
		Key:     dstKey,
		Account: src.Account,
		CondId:  src.ConditionId,
		Cond:    encodeCondition(src.Condition),
	}
}
