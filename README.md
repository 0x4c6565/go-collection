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
package main

import (
	"fmt"
	"github.com/0x4c6565/go-collection"
)

func main() {
	// Create a new collection from a slice
	numbers := collection.NewFromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	
	// Filter, transform, and compute
	result := numbers.
		Where(func(x int) bool { return x % 2 == 0 }). // Get even numbers
		Select(func(x int) any { return x * x }).      // Square each number
		Slice()                                        // Convert to slice
	
	fmt.Println(result) // Output: [4, 16, 36, 64, 100]
}
```

### Working with Structs

```go
package main

import (
	"fmt"
	"github.com/0x4c6565/go-collection"
)

type Person struct {
	Name string
	Age  int
}

func main() {
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
}
```

## Available Functions

### Collection Creation

- `New[T any, I iter.Seq[T] | []T](seq I) *Collection[T]` - Create a collection from an iterator or slice
- `NewFromIterator[T any](s iter.Seq[T]) *Collection[T]` - Create a collection from an iterator
- `NewFromSlice[T any](s []T) *Collection[T]` - Create a collection from a slice

### Filtering and Projection

- `Where(f func(x T) bool) *Collection[T]` - Filter elements based on a predicate
- `Select(f func(x T) any) *Collection[any]` - Transform elements using a selector function
- `SelectMany(f func(x T) *Collection[any]) *Collection[any]` - Project and flatten collections
- `Take(n int) *Collection[T]` - Get only the first n elements
- `Skip(n int) *Collection[T]` - Skip the first n elements
- `Distinct(equals func(a, b T) bool) *Collection[T]` - Get only distinct elements

### Ordering

- `OrderBy(f func(x T) any, ascending bool) *Collection[T]` - Order elements by a key
- `ThenBy(f func(x T) any, ascending bool) *Collection[T]` - Secondary ordering

### Element Operations

- `First() (T, bool)` - Get the first element or false
- `FirstOrError() (T, error)` - Get the first element or error
- `Last() (T, bool)` - Get the last element or false
- `LastOrError() (T, error)` - Get the last element or error

### Boolean Operations

- `All(f func(x T) bool) bool` - Check if all elements satisfy a condition
- `Any(f func(x T) bool) bool` - Check if any element satisfies a condition
- `Contains(f func(x T) bool) bool` - Check if collection contains elements satisfying a condition

### Set Operations

- `Union(other *Collection[T], equals func(a, b T) bool) *Collection[T]` - Union of two collections
- `Intersect(other *Collection[T], equals func(a, b T) bool) *Collection[T]` - Intersection of collections
- `Except(other *Collection[T], equals func(a, b T) bool) *Collection[T]` - Difference of collections
- `Concat(other *Collection[T]) *Collection[T]` - Concatenate collections

### Aggregation

- `Count() int` - Count elements in the collection
- `GroupBy(keySelector func(x T) any) map[any]*Collection[T]` - Group elements by key

### Numeric Operations

- `Average[T AverageTypes](c *Collection[T]) *big.Float` - Calculate average of numeric collection
- `Sum[T SumTypes](c *Collection[T]) *big.Float` - Calculate sum of numeric collection
- `SumInt[T SumIntTypes](c *Collection[T]) *big.Int` - Calculate sum of integer collection

### Conversion

- `Slice() []T` - Convert collection to a slice

## Error Handling

The package provides the `ErrNoElement` error which is returned when operations like `FirstOrError` or `LastOrError` are called on empty collections.