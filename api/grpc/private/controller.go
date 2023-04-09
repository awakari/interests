package private

import (
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
