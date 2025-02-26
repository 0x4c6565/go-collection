package gocollection

import (
	"iter"
	"slices"
)

type Collection[T any] iter.Seq[T]

func New[T any](seq iter.Seq[T]) *Collection[T] {
	d := Collection[T](seq)
	return &d
}

func NewFromSlice[T any](s []T) *Collection[T] {
	d := Collection[T](slices.Values(s))
	return &d
}

func (c *Collection[T]) First() T {
	for t := range *c {
		return t
	}

	panic("no items")
}

func (c *Collection[T]) Last() T {
	var l T
	for t := range *c {
		l = t
	}

	return l
}

func (c *Collection[T]) Count() int {
	count := 0
	for range *c {
		count++
	}

	return count
}

func (c *Collection[T]) Where(f func(x T) bool) *Collection[T] {
	return New(func(yield func(T) bool) {
		for v := range *c {
			if f(v) && !yield(v) {
				return
			}
		}
	})
}

func (c *Collection[T]) Contains(f func(x T) bool) bool {
	for t := range *c {
		if f(t) {
			return true
		}
	}
	return false
}

func (c *Collection[T]) Slice() []T {
	var val []T
	for t := range *c {
		val = append(val, t)
	}
	return val
}

func (c *Collection[T]) Select(f func(x T) any) *Collection[any] {
	return New(func(yield func(any) bool) {
		for v := range *c {
			if !yield(f(v)) {
				return
			}
		}
	})
}
