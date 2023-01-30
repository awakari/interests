package model

type GroupLogic int

const (
	GroupLogicOr = iota
	GroupLogicXor
	GroupLogicAnd
)

func (gl GroupLogic) String() string {
	return [...]string{
		"Or",
		"Xor",
		"And",
	}[gl]
}
