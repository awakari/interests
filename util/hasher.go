package util

type (
	HashCode uint64

	HashCoder interface {
		HashCode() HashCode
	}
)
