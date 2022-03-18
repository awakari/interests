package util

type HandleFunc[T interface{}] func(T) error

type HandleBatchFunc[T interface{}] func([]T) error
