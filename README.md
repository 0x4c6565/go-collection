# Collection Package for Go

A flexible and powerful library for working with collections in Go, inspired by LINQ-style operations. This package provides lazy evaluation and a fluent interface for manipulating collections of any type.

## Installation

```bash
go get github.com/0x4c6565/go-collection
```

## Overview

The `collection` package provides a generic way to work with collections of data using a fluent, method-chaining API. It supports lazy evaluation through iterators, allowing for efficient processing of large data sets.

## Example Usage

### Basic Filtering and Transformation

```go
// Create a new collection from a slice
numbers := collection.NewFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

// Filter, transform, and compute
result := numbers.
	Where(func(x int) bool { return x % 2 == 0 }). // Get even numbers
	Select(func(x int) any { return x * x }).      // Square each number
	ToSlice()                                        // Convert to slice

fmt.Println(result) // Output: [4, 16, 36, 64, 100]
```

### Working with Structs

```go
type Person struct {
	Name string
	Age  int
}

people := collection.NewFromSlice([]Person{
	{Name: "Alice", Age: 25},
	{Name: "Bob", Age: 30},
	{Name: "Charlie", Age: 22},
	{Name: "Lucy", Age: 17},
	{Name: "Dave", Age: 35},
})

// Find adults and order them by age
adults := people.
	Where(func(p Person) bool { return p.Age >= 18 }).
	OrderBy(func(p Person) any { return p.Age }, true) // Ascending order

for _, person := range *adults {
	fmt.Printf("%s: %d years old\n", person.Name, person.Age)
}

// Output:
// Charlie: 22 years old
// Alice: 25 years old
// Bob: 30 years old
// Dave: 35 years old
```

### Grouping and Aggregation

```go
type Product struct {
	Name     string
	Category string
	Price    float64
}

products := collection.NewFromSlice([]Product{
	{Name: "Apple", Category: "Fruit", Price: 1.99},
	{Name: "Banana", Category: "Fruit", Price: 0.99},
	{Name: "Carrot", Category: "Vegetable", Price: 0.50},
	{Name: "Potato", Category: "Vegetable", Price: 0.75},
})

// Group by category and calculate average price
groups := products.GroupBy(func(p Product) any { return p.Category })

for category, group := range groups {
	avg, _ := collection.AverageOrError(collection.Select(group, func(p Product) float64 { return p.Price }))
	fmt.Printf("Category: %s, Average Price: £%.2f\n", category, avg)
}

// Output:
// Category: Fruit, Average Price: £1.49
// Category: Vegetable, Average Price: £0.62
```

## Available Functions

### Collection Creation

- `New[T any, I iter.Seq[T] | []T](seq I) *Collection[T]` - Create a collection from an iterator or slice
- `NewFromIterator[T any](s iter.Seq[T]) *Collection[T]` - Create a collection from an iterator
- `NewFromSlice[T any](s []T) *Collection[T]` - Create a collection from a slice
- `NewFromItems[T any](s ...T) *Collection[T]` - Create a collection from given items
- `NewFromStringMap[T any](m map[string]T) *Collection[T]` - Create a collection from a string map
- `NewFromChannel[T any](ch <-chan T) *Collection[T]` - Create a collection from a channel
- `NewFromRange(start, count int) *Collection[int]` - Create a collection from a range of integers
- `NewFromJSON[T any](data []byte) (c *Collection[T], err error)` - Create a collection from a JSON string

### Filtering and Projection

- `Select[T any, e any](c *Collection[T], f func(x T) e) *Collection[e]` - Transform elements using a selector function
- `SelectMany[T any, E any](c *Collection[T], f func(x T) *Collection[E]) *Collection[E]` - Project and flatten collections

### Aggregation

- `Zip[T1, T2, TResult any](c1 *Collection[T1], c2 *Collection[T2], zipper func(T1, T2) TResult) *Collection[TResult]` - Combines two collections into one by applying a function pairwise
- `Join[TOuter, TInner, TKey comparable, TResult any](outer *Collection[TOuter], inner *Collection[TInner], outerKeySelector func(TOuter) TKey, innerKeySelector func(TInner) TKey, resultSelector func(TOuter, TInner) TResult) *Collection[TResult]` - Performs an inner join on two collections based on matching keys
- `Flatten[T any](c *Collection[*Collection[T]]) *Collection[T]` - Flattens a collection of collections into a single collection
- `Mode[T comparable](c *Collection[T]) (T, error)` - Return most frequently occurring element

### Conversion

- `ToMap[T any, K comparable](c *Collection[T], keySelector func(x T) K) map[K]T` - Converts a collection to a map

### Numeric Operations

- `AverageOrError[T NumericalTypes](c *Collection[T]) (*big.Float, error)` - Calculate average of numeric collection
- `Sum[T NumericalTypes](c *Collection[T]) *big.Float` - Calculate sum of numeric collection
- `Min[T NumericalTypes](c *Collection[T]) T` - Calculate the smallest value in the numeric collection
- `Max[T NumericalTypes](c *Collection[T]) T` - Calculate the largest value in the numeric collection
- `Median[T NumericalTypes](c *Collection[T]) (*big.Float, error)` - Calculate the median value in the numerical collection

## Available Collection Methods

### Filtering and Projection

- `Where(f func(x T) bool) *Collection[T]` - Filter elements by given predicate
- `Reject(f func(x T) bool) *Collection[T]` - Filter elements by given predicate
- `Find(f func(T) bool) (T, bool)` - Find first element by given predicate, returning boolean indicating whether found
- `Select(f func(x T) any) *Collection[any]` - Transform elements using a selector function
- `SelectMany(f func(x T) *Collection[any]) *Collection[any]` - Project and flatten collections
- `Take(n int) *Collection[T]` - Get only the first n elements
- `TakeUntil(f func(x T) bool) *Collection[T]` - Get elements until the predicate is satisfied
- `TakeWhile(f func(x T) bool) *Collection[T]` - Get elements whilst the predicate is satisfied
- `TakeLast(n int) *Collection[T]` - Take the last n elements
- `Skip(n int) *Collection[T]` - Skip the first n elements
- `SkipUntil(f func(x T) bool) *Collection[T]` - Skip elements until predicate is satisfied
- `SkipWhile(f func(x T) bool) *Collection[T]` - Skip elements whilst the predicate is satisfied
- `SkipLast(n int) *Collection[T]` - Skip the last n elements
- `Distinct(equals func(a, b T) bool) *Collection[T]` - Get only distinct elements

### Ordering

- `OrderBy(f func(x T) any, ascending bool) *Collection[T]` - Order elements by a key
- `Reverse() *Collection[T]` - Reverse elements
- `Shuffle() *Collection[T]` - Randomise elements

### Element Operations

- `First() (T, bool)` - Get the first element or false
- `FirstOrError() (T, error)` - Get the first element or error
- `Last() (T, bool)` - Get the last element or false
- `LastOrError() (T, error)` - Get the last element or error
- `ElementAt(index int) (T, bool)` - Get the element at index or false
- `ElementAtOrError(index int) (T, error)` - Get the element at index or error
- `IndexOf(predicate func(x T) bool) int` - Get the index of element that satisfies the predicate, or return `-1`
- `Partition(predicate func(x T) bool) (*Collection[T], *Collection[T])` - Divide collection into two based on predicate. The first collection contains elements that satisfy the predicate, the second contains elements that don't
- `ForEach(action func(v T))` - Execute action against each element. Consider iterating over collection instead
- `Each(action func(v T))` - Alias for ForEach()
- `ParallelForEach(ctx context.Context, action func(ctx context.Context, v T) error, concurrency int) error` - Execute action against each element in parallel
- `Peek(action func(T)) *Collection[T]` - Executes an action for each element in the collection and returns the collection

### Boolean Operations

- `All(f func(x T) bool) bool` - Check if all elements satisfy a condition
- `Any(f func(x T) bool) bool` - Check if any element satisfies a condition
- `Contains(f func(x T) bool) bool` - Check if collection contains elements satisfying a condition
- `IsEmpty() bool` - Returns boolean indicating if the collection is empty
- `Equals(other *Collection[T], equals func(a, b T) bool) bool` - compares collection with another to determine if they are equal

### Set Operations

- `Union(other *Collection[T], equals func(a, b T) bool) *Collection[T]` - Union of two collections
- `Intersect(other *Collection[T], equals func(a, b T) bool) *Collection[T]` - Intersection of collections
- `Except(other *Collection[T], equals func(a, b T) bool) *Collection[T]` - Difference of collections
- `Concat(other *Collection[T]) *Collection[T]` - Concatenate collections
- `Append(e T) *Collection[T]` - Add element to the end of the collection
- `Prepend(e T) *Collection[T]` - Add element to the beginning of the collection
- `Pop() (d T, err error)` - Removes the last element from collection and returns it
- `Shift() (d T, err error)` - Removes the first element from collection and returns it

### Aggregation

- `Len() int` - Number of elements in the collection
- `Count() int` - Alias for Len()
- `GroupBy(keySelector func(x T) any) map[any]*Collection[T]` - Group elements by key
- `Chunk(size int) []*Collection[T]` Split collection into chunks of the specified size
- `Aggregate(seed any, accumulator func(result any, item T) any) any` - Applies an accumulator function over collection

### Conversion

- `ToSlice() []T` - Convert collection to a slice
- `ToMap(keySelector func(x T) any) map[any]T` - Convert collection to a map
- `ToChannel() <-chan T` - Convert collection to a channel
- `ToJSON() ([]byte, error)` - Serialise collection into JSON string

## Errors

- `ErrNoElement` - Returned when methods like `FirstOrError` or `LastOrError` are called on empty collections
- `ErrIndexOutOfRange` - Returned when methods like `ElementAtOrError` are called with out of bound indexes
