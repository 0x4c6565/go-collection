package collection

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"math/big"
	"math/rand"
	"runtime"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"
)

var ErrNoElement = errors.New("no element")
var ErrIndexOutOfRange = errors.New("index out of range")
var ErrEmptyCollection = errors.New("empty collection")
var ErrNotExactlyOneElement = errors.New("not exactly one element")

type Collection[T any] func(yield func(T) bool)

// New creates a new Collection from either an iterator or a slice
func New[T any, I iter.Seq[T] | []T](seq I) *Collection[T] {
	if s, ok := any(seq).([]T); ok {
		return NewFromSlice(s)
	} else {
		return NewFromIterator(any(seq).(iter.Seq[T]))
	}
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

// NewFromItems creates a new Collection from given items
func NewFromItems[T any](s ...T) *Collection[T] {
	d := Collection[T](slices.Values(s))
	return &d
}

// NewFromMap creates a new Collection from a map with string keys
func NewFromMap[K comparable, V any](m map[K]V) *Collection[V] {
	var values []V
	for _, v := range m {
		values = append(values, v)
	}
	return NewFromSlice(values)
}

// NewFromChannel creates a new Collection from a channel
func NewFromChannel[T any](ch <-chan T) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	}))
}

// NewFromRange creates a new Collection from a range of integers
func NewFromRange(start, count int) *Collection[int] {
	if count < 0 {
		return NewFromSlice([]int{})
	}

	return New[int](iter.Seq[int](func(yield func(int) bool) {
		for i := 0; i < count; i++ {
			if !yield(start + i) {
				return
			}
		}
	}))
}

// NewFromJSON deserializes JSON into a new collection
func NewFromJSON[T any](data []byte) (c *Collection[T], err error) {
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return c, fmt.Errorf("failed to unmarshal Collection: %w", err)
	}
	c = NewFromSlice(items)
	return
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

// Reject filters the collection to only elements not satisfying the predicate function
func (c *Collection[T]) Reject(f func(x T) bool) *Collection[T] {
	return c.Where(func(x T) bool {
		return !f(x)
	})
}

// Find returns the first element that matches the given predicate.
// If no element matches, it returns the zero value and false.
func (c *Collection[T]) Find(f func(T) bool) (v T, ok bool) {
	for v := range *c {
		if f(v) {
			return v, true
		}
	}
	return
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

// Single returns the only element in the collection and a boolean indicating if an element was found
func (c *Collection[T]) Single() (element T, ok bool) {
	count := 0
	for v := range *c {
		if count > 0 {
			return
		}
		element = v
		count++
	}

	if count == 1 {
		return element, true
	}

	return
}

// SingleOrError returns the only element in the collection or an error if not exactly one element
func (c *Collection[T]) SingleOrError() (element T, err error) {
	element, ok := c.Single()
	if !ok {
		return element, ErrNotExactlyOneElement
	}

	return
}

// Len returns the number of elements in the collection
func (c *Collection[T]) Len() int {
	count := 0
	for range *c {
		count++
	}

	return count
}

// Count is an alias for Len
func (c *Collection[T]) Count() int { return c.Len() }

// Contains returns true if any element satisfies the predicate
func (c *Collection[T]) Contains(f func(x T) bool) bool {
	for t := range *c {
		if f(t) {
			return true
		}
	}
	return false
}

// IsEmpty returns true if the collection is empty
func (c *Collection[T]) IsEmpty() bool {
	for range *c {
		return false
	}
	return true
}

func (c *Collection[T]) Shuffle() *Collection[T] {
	slice := c.ToSlice()
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	return NewFromSlice(slice)
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
		slice := c.ToSlice()
		for i := range len(slice) - n {
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
		slice := c.ToSlice()

		start := max(len(slice)-n, 0)

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

// None returns true if no elements satisfy the predicate
func (c *Collection[T]) None(f func(x T) bool) bool {
	return !c.Any(f)
}

type NumericalTypes interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
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
	slice := c.ToSlice()

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
			current := group.ToSlice()
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
			for v2 := range *other {
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
			for v2 := range *other {
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

// Equals compares collection with another to determine if they are equal
func (c *Collection[T]) Equals(other *Collection[T], equals func(a, b T) bool) bool {
	iter1 := c.ToSlice()
	iter2 := other.ToSlice()

	if len(iter1) != len(iter2) {
		return false
	}

	for i := range iter1 {
		if !equals(iter1[i], iter2[i]) {
			return false
		}
	}

	return true
}

// Reverse returns a collection with the elements in reverse order
func (c *Collection[T]) Reverse() *Collection[T] {
	slice := c.ToSlice()
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

// ForEach executes an action for each element in the collection
func (c *Collection[T]) ForEach(action func(v T)) {
	for v := range *c {
		action(v)
	}
}

// Each is an alias for ForEach
func (c *Collection[T]) Each(action func(v T)) { c.ForEach(action) }

// ParallelForEach executes an action for each element in the collection in parallel
func (c *Collection[T]) ParallelForEach(ctx context.Context, action func(ctx context.Context, v T) error, concurrency int) error {
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)
	for item := range *c {
		currentItem := item
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return action(ctx, currentItem)
			}
		})
	}

	return g.Wait()
}

// Peek executes an action for each element in the collection and returns the collection
func (c *Collection[T]) Peek(action func(T)) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for v := range *c {
			action(v)
			if !yield(v) {
				return
			}
		}
	}))
}

// ElementAt returns the element at the specified index or a default value if index is out of range
func (c *Collection[T]) ElementAt(index int) (v T, ok bool) {
	if index < 0 {
		return
	}

	count := 0
	for v := range *c {
		if count == index {
			return v, true
		}
		count++
	}
	return
}

// ElementAtOrError returns the element at the specified index or an error if index is out of range
func (c *Collection[T]) ElementAtOrError(index int) (T, error) {
	val, ok := c.ElementAt(index)
	if !ok {
		return val, ErrIndexOutOfRange
	}
	return val, nil
}

// Random returns a random element from the collection and true, or a default value and false if collection is empty
func (c *Collection[T]) Random() (v T, ok bool) {
	slice := c.ToSlice()
	if len(slice) == 0 {
		return
	}

	i, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(slice))))
	if err != nil {
		return
	}

	return slice[i.Int64()], true
}

// RandomN returns n random elements from the collection and true, or an empty slice and false if not enough elements
func (c *Collection[T]) RandomN(n int) (v []T, ok bool) {
	slice := c.ToSlice()
	if n <= 1 || len(slice) < n {
		return
	}

	for range n {
		i, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(slice))))
		if err != nil {
			return
		}
		v = append(v, slice[i.Int64()])
	}

	return v, true
}

// IndexOf returns the index of the first element that satisfies the predicate
func (c *Collection[T]) IndexOf(predicate func(x T) bool) int {
	index := 0
	for item := range *c {
		if predicate(item) {
			return index
		}
		index++
	}
	return -1
}

// Partition divides the collection into two collections based on a predicate function.
// The first collection contains elements that satisfy the predicate, the second contains elements that don't.
func (c *Collection[T]) Partition(predicate func(x T) bool) (*Collection[T], *Collection[T]) {
	var matches []T
	var nonMatches []T

	// Pre-allocate slices to improve performance for large collections
	// Allocate with a conservative initial capacity
	initialCapacity := c.Len() / 2
	if initialCapacity > 0 {
		matches = make([]T, 0, initialCapacity)
		nonMatches = make([]T, 0, initialCapacity)
	}

	for v := range *c {
		if predicate(v) {
			matches = append(matches, v)
		} else {
			nonMatches = append(nonMatches, v)
		}
	}

	return NewFromSlice(matches), NewFromSlice(nonMatches)
}

// ToSlice converts the collection to a slice
func (c *Collection[T]) ToSlice() []T {
	var val []T
	for t := range *c {
		val = append(val, t)
	}
	return val
}

// ToMap converts the collection to a map with string keys
func (c *Collection[T]) ToMap(keySelector func(x T) any) map[any]T {
	return ToMap(c, keySelector)
}

// ToChannel converts the collection to a channel
func (c *Collection[T]) ToChannel() <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for item := range *c {
			ch <- item
		}
	}()
	return ch
}

// ToJSON serializes the collection to JSON
func (c *Collection[T]) ToJSON() ([]byte, error) {
	return json.Marshal(c.ToSlice())
}

// Pop removes the last element from collection and returns it
func (c *Collection[T]) Pop() (v T, err error) {
	s := c.ToSlice()
	if len(s) == 0 {
		return v, ErrEmptyCollection
	}
	last := s[len(s)-1]
	*c = *NewFromSlice(s[:len(s)-1])
	return last, nil
}

// Shift removes the first element from collection and returns it
func (c *Collection[T]) Shift() (v T, err error) {
	s := c.ToSlice()
	if len(s) == 0 {
		return v, ErrEmptyCollection
	}
	first := s[0]
	*c = *NewFromSlice(s[1:])
	return first, nil
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

// Join performs an inner join on two collections based on matching keys
func Join[TOuter, TInner, TKey comparable, TResult any](outer *Collection[TOuter], inner *Collection[TInner], outerKeySelector func(TOuter) TKey, innerKeySelector func(TInner) TKey, resultSelector func(TOuter, TInner) TResult) *Collection[TResult] {
	return New[TResult](iter.Seq[TResult](func(yield func(TResult) bool) {
		for outerItem := range *outer {
			outerKey := outerKeySelector(outerItem)

			for innerItem := range *inner {
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

// Flatten flattens a collection of collections into a single collection
func Flatten[T any](c *Collection[*Collection[T]]) *Collection[T] {
	return New[T](iter.Seq[T](func(yield func(T) bool) {
		for innerCollection := range *c {
			for v := range *innerCollection {
				if !yield(v) {
					return
				}
			}
		}
	}))
}

// Mode returns the most frequently occurring element in the collection.
// If multiple values have the same frequency, the first one is returned
func Mode[T comparable](c *Collection[T]) (mode T, err error) {
	slice := c.ToSlice()
	if len(slice) == 0 {
		return mode, ErrEmptyCollection
	}

	freq := make(map[T]int)
	var keys []T
	for _, v := range slice {
		if _, ok := freq[v]; !ok {
			keys = append(keys, v)
		}
		freq[v]++
	}

	var maxCount int
	for _, key := range keys {
		count := freq[key]
		if count > maxCount {
			mode = key
			maxCount = count
		}
	}

	return
}

// Map converts the collection to a map
func ToMap[T any, K comparable](c *Collection[T], keySelector func(x T) K) map[K]T {
	m := make(map[K]T)
	for v := range *c {
		m[keySelector(v)] = v
	}
	return m
}

// AverageOrError calculates the average or returns an error if empty
func AverageOrError[T NumericalTypes](c *Collection[T]) (*big.Float, error) {
	sum := float64(0)
	count := 0
	for t := range *c {
		sum += float64(t)
		count++
	}
	if count == 0 {
		return nil, errors.New("cannot compute average of empty collection")
	}
	return big.NewFloat(sum / float64(count)), nil
}

// Sum calculates the sum of all elements in the collection and returns it as a big.Float
func Sum[T NumericalTypes](c *Collection[T]) *big.Float {
	sum := float64(0)
	for t := range *c {
		sum += float64(t)
	}
	return big.NewFloat(sum)
}

// Min returns the smallest value in the collection
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

// Median calculates the median of the collection
func Median[T NumericalTypes](c *Collection[T]) (*big.Float, error) {
	slice := c.ToSlice()
	if len(slice) == 0 {
		return nil, errors.New("cannot compute median of empty collection")
	}

	slices.Sort(slice)

	mid := len(slice) / 2
	if len(slice)%2 == 0 {
		return big.NewFloat(float64(slice[mid-1]+slice[mid]) / 2), nil
	}
	return big.NewFloat(float64(slice[mid])), nil
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
