package subscription

type QueryOwn struct {
	Limit   uint32
	GroupId string
	UserId  string
}

type QueryByCondition struct {
	CondId string
	Limit  uint32
}
