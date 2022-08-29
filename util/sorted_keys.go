package util

import (
	"sort"
)

func SortedKeys[K Ordered, V any](m map[K]V) (keys []K) {
	keys = make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return
}
