package collection

import (
	"context"
	"errors"
	"iter"
	"math/big"
	"runtime"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"
)

var ErrNoElement = errors.New("no element")
var ErrIndexOutOfRange = errors.New("index out of range")

type Collection[T any] func(yield func(T) bool)

// New creates a new Collection from either a iterator or a slice
func New[T any, I iter.Seq[T] | []T](seq I) *Collection[T] {
	var d Collection[T]
	if s, ok := any(seq).([]T); ok {
		d = Collection[T](slices.Values(s))
	} else {
		d = Collection[T](any(seq).(iter.Seq[T]))
	}
	return &d
}

// NewFromIterator creates a new Collection from an iterator
func NewFromIterator[T any](s iter.Seq[T]) *Collection[T] {
	d := Collection[T](s)
	return &d
}

// NewFromSlice creates a new Collection from a slice
func NewFromSlice[T any](s []T) *Collection[T] {
	d := Collection[T](slices.Values(s))
	return &d
}

// NewFromStringMap creates a new Collection from a map with string keys
func NewFromStringMap[T any](m map[string]T) *Collection[T] {
	var values []T
	for _, v := range m {
		values = append(values, v)
	}
	return NewFromSlice(values)
}

// Where filters the collection to only elements satisfying the predicate function
func (c *Collection[T]) Where(f func(x T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range *c {
			if f(v) && !yield(v) {
				return
			}
		}
	}))
}

// Select transforms each element in the collection using the selector function
func (c *Collection[T]) Select(f func(x T) any) *Collection[any] {
	return Select(c, f)
}

// SelectMany projects each element of the collection to a new collection and flattens the resulting collections into one
func (c *Collection[T]) SelectMany(f func(x T) *Collection[any]) *Collection[any] {
	return SelectMany(c, f)
}

// All returns true if all elements satisfy the predicate
func (c *Collection[T]) All(f func(x T) bool) bool {
	for t := range *c {
		if !f(t) {
			return false
		}
	}
	return true
}

// First returns the first element in the collection and a boolean indicating if an element was found
func (c *Collection[T]) First() (first T, ok bool) {
	for t := range *c {
		return t, true
	}

	return first, false
}

// FirstOrError returns the first element or an error if the collection is empty
func (c *Collection[T]) FirstOrError() (first T, err error) {
	first, ok := c.First()
	if !ok {
		return first, ErrNoElement
	}

	return
}

// Last returns the last element in the collection and a boolean indicating if an element was found
func (c *Collection[T]) Last() (last T, ok bool) {
	for t := range *c {
		last = t
		ok = true
	}

	return
}

// LastOrError returns the last element or an error if the collection is empty
func (c *Collection[T]) LastOrError() (last T, err error) {
	last, ok := c.Last()
	if !ok {
		return last, ErrNoElement
	}

	return
}

// Count returns the number of elements in the collection
func (c *Collection[T]) Count() int {
	count := 0
	for range *c {
		count++
	}

	return count
}

// Contains returns true if any element satisfies the predicate
func (c *Collection[T]) Contains(f func(x T) bool) bool {
	for t := range *c {
		if f(t) {
			return true
		}
	}
	return false
}

// Distinct returns a collection containing only distinct elements based on the provided equality function
func (c *Collection[T]) Distinct(equals func(a, b T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		seen := make([]T, 0)
		for v := range *c {
			unique := true
			for _, s := range seen {
				if equals(v, s) {
					unique = false
					break
				}
			}
			if unique {
				seen = append(seen, v)
				if !yield(v) {
					return
				}
			}
		}
	}))
}

// Skip returns a collection that skips the first n elements
func (c *Collection[T]) Skip(n int) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		count := 0
		for v := range *c {
			if count >= n {
				if !yield(v) {
					return
				}
			}
			count++
		}
	}))
}

// SkipUntil returns a collection that skips elements until the predicate is satisfied
func (c *Collection[T]) SkipUntil(f func(x T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		skip := true
		for v := range *c {
			if !skip || f(v) {
				skip = false
				if !yield(v) {
					return
				}
			}
		}
	}))
}

func (c *Collection[T]) SkipWhile(f func(x T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		skip := true
		for v := range *c {
			if !skip || !f(v) {
				skip = false
				if !yield(v) {
					return
				}
			}
		}
	}))
}

// SkipLast returns a collection that skips the last n elements
func (c *Collection[T]) SkipLast(n int) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		slice := c.Slice()
		for i := 0; i < len(slice)-n; i++ {
			if !yield(slice[i]) {
				return
			}
		}
	}))
}

// Take returns a collection of only the first n elements
func (c *Collection[T]) Take(n int) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		count := 0
		for v := range *c {
			if count < n {
				if !yield(v) {
					return
				}
				count++
			} else {
				return
			}
		}
	}))
}

// TakeUntil returns a collection of elements until the predicate is satisfied
func (c *Collection[T]) TakeUntil(f func(x T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range *c {
			if f(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}))
}

func (c *Collection[T]) TakeWhile(f func(x T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range *c {
			if !f(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}))
}

// TakeLast returns a collection of only the last n elements
func (c *Collection[T]) TakeLast(n int) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		slice := c.Slice()

		start := len(slice) - n
		if start < 0 {
			start = 0
		}

		for i := start; i < len(slice); i++ {
			if !yield(slice[i]) {
				return
			}
		}
	}))
}

// Any returns true if any element satisfies the predicate
func (c *Collection[T]) Any(f func(x T) bool) bool {
	for t := range *c {
		if f(t) {
			return true
		}
	}
	return false
}

type NumericalTypes interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64 | float32 | float64
}

func orderByNumerical[T NumericalTypes](a T, b T, ascending bool) int {
	if ascending {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}
	if a > b {
		return -1
	} else if a < b {
		return 1
	}
	return 0
}

// OrderBy returns a collection ordered by the key selector
func (c *Collection[T]) OrderBy(f func(x T) any, ascending bool) *Collection[T] {
	slice := c.Slice()

	slices.SortFunc(slice, func(a, b T) int {
		aValue, bValue := f(a), f(b)

		switch aValueTyped := aValue.(type) {
		case int:
			return orderByNumerical(aValueTyped, bValue.(int), ascending)
		case int8:
			return orderByNumerical(aValueTyped, bValue.(int8), ascending)
		case int16:
			return orderByNumerical(aValueTyped, bValue.(int16), ascending)
		case int32:
			return orderByNumerical(aValueTyped, bValue.(int32), ascending)
		case int64:
			return orderByNumerical(aValueTyped, bValue.(int64), ascending)
		case uint:
			return orderByNumerical(aValueTyped, bValue.(uint), ascending)
		case uint8:
			return orderByNumerical(aValueTyped, bValue.(uint8), ascending)
		case uint16:
			return orderByNumerical(aValueTyped, bValue.(uint16), ascending)
		case uint32:
			return orderByNumerical(aValueTyped, bValue.(uint32), ascending)
		case uint64:
			return orderByNumerical(aValueTyped, bValue.(uint64), ascending)
		case float32:
			return orderByNumerical(aValueTyped, bValue.(float32), ascending)
		case float64:
			return orderByNumerical(aValueTyped, bValue.(float64), ascending)
		case string:
			bValueTyped := bValue.(string)
			if ascending {
				return strings.Compare(aValueTyped, bValueTyped)
			}
			return strings.Compare(bValueTyped, aValueTyped)
		default:
			// For other types, use basic comparison
			return 0
		}
	})

	return NewFromSlice(slice)
}

// Concat combines two collections into one
func (c *Collection[T]) Concat(other *Collection[T]) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range *c {
			if !yield(v) {
				return
			}
		}

		for v := range *other {
			if !yield(v) {
				return
			}
		}
	}))
}

// GroupBy groups elements by a key selector
func (c *Collection[T]) GroupBy(keySelector func(x T) any) map[any]*Collection[T] {
	groups := make(map[any]*Collection[T])

	for v := range *c {
		key := keySelector(v)
		if group, exists := groups[key]; exists {
			// Add to existing group
			current := group.Slice()
			current = append(current, v)
			groups[key] = NewFromSlice(current)
		} else {
			// Create new group
			groups[key] = NewFromSlice([]T{v})
		}
	}

	return groups
}

// Union returns a collection of distinct elements from both collections
func (c *Collection[T]) Union(other *Collection[T], equals func(a, b T) bool) *Collection[T] {
	return c.Concat(other).Distinct(equals)
}

// Intersect returns a collection of elements present in both collections
func (c *Collection[T]) Intersect(other *Collection[T], equals func(a, b T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v1 := range *c {
			exists := false
			otherSlice := other.Slice()
			for _, v2 := range otherSlice {
				if equals(v1, v2) {
					exists = true
					break
				}
			}

			if exists {
				if !yield(v1) {
					return
				}
			}
		}
	}))
}

// Except returns a collection of elements in this collection but not in the other
func (c *Collection[T]) Except(other *Collection[T], equals func(a, b T) bool) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v1 := range *c {
			exists := false
			otherSlice := other.Slice()
			for _, v2 := range otherSlice {
				if equals(v1, v2) {
					exists = true
					break
				}
			}

			if !exists {
				if !yield(v1) {
					return
				}
			}
		}
	}))
}

// Reverse returns a collection with the elements in reverse order
func (c *Collection[T]) Reverse() *Collection[T] {
	slice := c.Slice()
	slices.Reverse(slice)
	return NewFromSlice(slice)
}

// Append adds an element to the end of the collection
func (c *Collection[T]) Append(e T) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range *c {
			if !yield(v) {
				return
			}
		}
		if !yield(e) {
			return
		}
	}))
}

// Prepend adds an element to the beginning of the collection
func (c *Collection[T]) Prepend(e T) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		if !yield(e) {
			return
		}
		for v := range *c {
			if !yield(v) {
				return
			}
		}
	}))
}

// Chunk splits the collection into chunks of the specified size
func (c *Collection[T]) Chunk(size int) []*Collection[T] {
	chunks := make([]*Collection[T], 0)
	chunk := make([]T, 0)

	for v := range *c {
		chunk = append(chunk, v)
		if len(chunk) == size {
			chunks = append(chunks, NewFromSlice(chunk))
			chunk = make([]T, 0)
		}
	}

	if len(chunk) > 0 {
		chunks = append(chunks, NewFromSlice(chunk))
	}

	return chunks
}

// Aggregate applies an accumulator function over collection
func (c *Collection[T]) Aggregate(seed any, accumulator func(result any, item T) any) any {
	result := seed
	for item := range *c {
		result = accumulator(result, item)
	}
	return result
}

// ParallelForEach executes an action for each element in the collection in parallel
func (c *Collection[T]) ParallelForEach(ctx context.Context, action func(T) error, numWorkers int) error {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(numWorkers)
	slice := c.Slice()
	for _, item := range slice {
		currentItem := item
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return action(currentItem)
			}
		})
	}

	return g.Wait()
}

// Zip combines two collections into one by applying a function pairwise
func Zip[T1, T2, TResult any](c1 *Collection[T1], c2 *Collection[T2], zipper func(T1, T2) TResult) *Collection[TResult] {
	return New[TResult](iter.Seq[TResult](func(yield func(TResult) bool) {
		iter1 := make(chan T1)
		iter2 := make(chan T2)

		// Start goroutines to generate values
		go func() {
			defer close(iter1)
			for v := range *c1 {
				iter1 <- v
			}
		}()

		go func() {
			defer close(iter2)
			for v := range *c2 {
				iter2 <- v
			}
		}()

		// Zip elements together
		for {
			v1, ok1 := <-iter1
			if !ok1 {
				break
			}

			v2, ok2 := <-iter2
			if !ok2 {
				break
			}

			if !yield(zipper(v1, v2)) {
				return
			}
		}
	}))
}

func (c *Collection[T]) ElementAt(index int) (T, bool) {
	var d T
	if index < 0 {
		return d, false
	}

	count := 0
	for v := range *c {
		if count == index {
			return v, true
		}
		count++
	}
	return d, false
}

func (c *Collection[T]) ElementAtOrError(index int) (T, error) {
	val, ok := c.ElementAt(index)
	if !ok {
		return val, ErrIndexOutOfRange
	}
	return val, nil
}

// Join performs an inner join on two collections based on matching keys
func Join[TOuter, TInner, TKey comparable, TResult any](outer *Collection[TOuter], inner *Collection[TInner], outerKeySelector func(TOuter) TKey, innerKeySelector func(TInner) TKey, resultSelector func(TOuter, TInner) TResult) *Collection[TResult] {
	return New[TResult](iter.Seq[TResult](func(yield func(TResult) bool) {
		innerSlice := inner.Slice()

		for outerItem := range *outer {
			outerKey := outerKeySelector(outerItem)

			for _, innerItem := range innerSlice {
				innerKey := innerKeySelector(innerItem)

				if outerKey == innerKey {
					if !yield(resultSelector(outerItem, innerItem)) {
						return
					}
				}
			}
		}
	}))
}

// Slice converts the collection to a slice
func (c *Collection[T]) Slice() []T {
	var val []T
	for t := range *c {
		val = append(val, t)
	}
	return val
}

// StringMap converts the collection to a map with string keys
func (c *Collection[T]) StringMap(keySelector func(x T) string) map[string]T {
	m := make(map[string]T)
	for v := range *c {
		m[keySelector(v)] = v
	}
	return m
}

// Average calculates the average value of a numeric collection
func Average[T NumericalTypes](c *Collection[T]) *big.Float {
	sum := float64(0)
	count := 0
	for t := range *c {
		sum += float64(t)
		count++
	}
	return big.NewFloat(sum / float64(count))
}

// Sum calculates the sum of all elements in the collection and returns it as a big.Float
func Sum[T NumericalTypes](c *Collection[T]) *big.Float {
	sum := float64(0)
	for t := range *c {
		sum += float64(t)
	}
	return big.NewFloat(sum)
}

func Min[T NumericalTypes](c *Collection[T]) T {
	min := T(0)
	first := true
	for t := range *c {
		if first || t < min {
			min = t
		}
		first = false
	}
	return min
}

// Max returns the largest value in the collection
func Max[T NumericalTypes](c *Collection[T]) T {
	max := T(0)
	first := true
	for t := range *c {
		if first || t > max {
			max = t
		}
		first = false
	}
	return max
}

// Select transforms each element in the collection using the selector function
func Select[T any, E any](c *Collection[T], f func(x T) E) *Collection[E] {
	return New[E](iter.Seq[E](func(yield func(E) bool) {
		for v := range *c {
			if !yield(f(v)) {
				return
			}
		}
	}))
}

// SelectMany projects each element of the collection to a new collection and flattens the resulting collections into one
func SelectMany[T any, E any](c *Collection[T], f func(x T) *Collection[E]) *Collection[E] {
	return New[E](iter.Seq[E](func(yield func(E) bool) {
		for v := range *c {
			innerCollection := f(v)
			for innerValue := range *innerCollection {
				if !yield(innerValue) {
					return
				}
			}
		}
	}))
}
