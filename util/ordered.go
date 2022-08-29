package util

import "golang.org/x/exp/constraints"

type Ordered interface {
	constraints.Ordered | ~byte | ~string
}
