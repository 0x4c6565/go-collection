package gocollection

import (
	"iter"
	"slices"
)

type GoCollection[T any] iter.Seq[T]

func New[T any](seq iter.Seq[T]) *GoCollection[T] {
	d := GoCollection[T](seq)
	return &d
}

func NewFromSlice[T any](s []T) *GoCollection[T] {
	d := GoCollection[T](slices.Values(s))
	return &d
}

func (c *GoCollection[T]) First() T {
	for t := range *c {
		return t
	}

	panic("no items")
}

func (c *GoCollection[T]) Last() T {
	var l T
	for t := range *c {
		l = t
	}

	return l
}

func (c *GoCollection[T]) Count() int {
	count := 0
	for range *c {
		count++
	}

	return count
}

func (c *GoCollection[T]) Where(f func(x T) bool) *GoCollection[T] {
	return New(func(yield func(T) bool) {
		for v := range *c {
			if f(v) && !yield(v) {
				return
			}
		}
	})
}

func (c *GoCollection[T]) Contains(f func(x T) bool) bool {
	for t := range *c {
		if f(t) {
			return true
		}
	}
	return false
}

func (c *GoCollection[T]) Slice() []T {
	var val []T
	for t := range *c {
		val = append(val, t)
	}
	return val
}

func (c *GoCollection[T]) Select(f func(x T) any) *GoCollection[any] {
	return New(func(yield func(any) bool) {
		for v := range *c {
			if !yield(f(v)) {
				return
			}
		}
	})
}
