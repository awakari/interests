package subscription

type QueryOwn struct {
	Limit   uint32
	GroupId string
	UserId  string
	Order   Order
	Pattern string
}

type QueryByCondition struct {
	CondId string
	Limit  uint32
}

type Order int

const (
	OrderAsc Order = iota
	OrderDesc
)

func (o Order) String() string {
	return [...]string{
		"Asc",
		"Desc",
	}[o]
}
