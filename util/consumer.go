package util

type ConsumeFunc[T any] func(item T) (err error)
