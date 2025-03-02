package collection_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	collection "github.com/0x4c6565/go-collection"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Slice", func(t *testing.T) {
		c := collection.New[string]([]string{"a", "b", "c"})
		v, _ := c.First()

		assert.Equal(t, "a", v)
	})
	t.Run("Iterator", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		c := collection.New[string](slices.Values(s))
		v, _ := c.First()

		assert.Equal(t, "a", v)
	})
}

func TestNewFromSlice(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	v, _ := c.First()

	assert.Equal(t, "a", v)
}

func TestNewFromIterator(t *testing.T) {
	s := []string{"a", "b", "c"}
	c := collection.NewFromIterator(slices.Values(s))
	v, _ := c.First()

	assert.Equal(t, "a", v)
}

func TestWhere(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	t.Run("Elements", func(t *testing.T) {
		v := c.Where(func(x string) bool {
			return x == "a"
		}).Slice()

		assert.Len(t, v, 1)
		assert.Equal(t, "a", v[0])
	})

	t.Run("NoElements", func(t *testing.T) {
		v := c.Where(func(x string) bool {
			return x == "z"
		}).Slice()

		assert.Len(t, v, 0)
	})

	t.Run("Break", func(t *testing.T) {
		for range *c.Where(func(x string) bool {
			return x == "a"
		}) {
			break
		}
	})
}

func TestSelect(t *testing.T) {
	type teststruct struct {
		Property1 string
		Property2 int
	}
	var teststructs []teststruct
	teststructs = append(teststructs, teststruct{
		Property1: "s1",
		Property2: 1,
	}, teststruct{
		Property1: "s2",
		Property2: 2,
	})
	c := collection.New[teststruct](teststructs)

	t.Run("ReturnProperty", func(t *testing.T) {
		var results []int
		for _, val := range c.Select(func(x teststruct) any {
			return x.Property2
		}).Slice() {
			results = append(results, val.(int))
		}

		assert.Equal(t, 1, results[0])
		assert.Equal(t, 2, results[1])
	})

	t.Run("Break", func(t *testing.T) {
		for range *c.Select(func(x teststruct) any {
			return x.Property2
		}) {
			break
		}
	})
}

func TestSelectMany(t *testing.T) {
	type teststruct struct {
		Property1 string
		Property2 int
		Children  []string
	}

	var teststructs []teststruct
	teststructs = append(teststructs, teststruct{
		Property1: "s1",
		Property2: 1,
		Children:  []string{"child1", "child2"},
	}, teststruct{
		Property1: "s2",
		Property2: 2,
		Children:  []string{"child3", "child4", "child5"},
	})

	c := collection.New[teststruct](teststructs)

	t.Run("FlattenChildren", func(t *testing.T) {
		var results []string
		for _, val := range c.SelectMany(func(x teststruct) *collection.Collection[any] {
			childrenAny := make([]any, len(x.Children))
			for i, child := range x.Children {
				childrenAny[i] = child
			}
			return collection.New[any](childrenAny)
		}).Slice() {
			results = append(results, val.(string))
		}

		assert.Equal(t, 5, len(results))
		assert.Equal(t, "child1", results[0])
		assert.Equal(t, "child2", results[1])
		assert.Equal(t, "child3", results[2])
		assert.Equal(t, "child4", results[3])
		assert.Equal(t, "child5", results[4])
	})

	t.Run("Break", func(t *testing.T) {
		for range *c.SelectMany(func(x teststruct) *collection.Collection[any] {
			childrenAny := make([]any, len(x.Children))
			for i, child := range x.Children {
				childrenAny[i] = child
			}
			return collection.New[any](childrenAny)
		}) {
			break
		}
	})

	t.Run("EmptyChildren", func(t *testing.T) {
		var emptyTeststructs []teststruct
		emptyTeststructs = append(emptyTeststructs, teststruct{
			Property1: "s1",
			Property2: 1,
			Children:  []string{},
		}, teststruct{
			Property1: "s2",
			Property2: 2,
			Children:  []string{},
		})

		emptyC := collection.New[teststruct](emptyTeststructs)

		var results []string
		for _, val := range emptyC.SelectMany(func(x teststruct) *collection.Collection[any] {
			childrenAny := make([]any, len(x.Children))
			for i, child := range x.Children {
				childrenAny[i] = child
			}
			return collection.New[any](childrenAny)
		}).Slice() {
			results = append(results, val.(string))
		}

		assert.Equal(t, 0, len(results))
	})

	t.Run("MixedTypes", func(t *testing.T) {
		var results []any
		results = append(results, c.SelectMany(func(x teststruct) *collection.Collection[any] {
			return collection.New[any]([]any{x.Property1, x.Property2})
		}).Slice()...)

		assert.Equal(t, 4, len(results))
		assert.Equal(t, "s1", results[0])
		assert.Equal(t, 1, results[1])
		assert.Equal(t, "s2", results[2])
		assert.Equal(t, 2, results[3])
	})
}

func TestAll(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v := c.All(func(x string) bool {
			return x != ""
		})

		assert.True(t, v)
	})

	t.Run("False", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v := c.All(func(x string) bool {
			return x == "a"
		})

		assert.False(t, v)
	})
}

func TestFirst(t *testing.T) {
	t.Run("NoElement_OkFalse", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})
		_, ok := c.First()

		assert.False(t, ok)
	})
	t.Run("Element_OkTrue", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v, ok := c.First()

		assert.True(t, ok)
		assert.Equal(t, "a", v)
	})
}

func TestFirstOrError(t *testing.T) {
	t.Run("Element", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v, err := c.FirstOrError()

		assert.Nil(t, err)
		assert.Equal(t, "a", v)
	})

	t.Run("NoElement_Error", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})
		_, err := c.FirstOrError()

		assert.NotNil(t, err)
		assert.Equal(t, collection.ErrNoElement, err)
	})
}

func TestLast(t *testing.T) {
	t.Run("NoElement_OkFalse", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})
		_, ok := c.Last()

		assert.False(t, ok)
	})
	t.Run("Element_OkTrue", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v, ok := c.Last()

		assert.True(t, ok)
		assert.Equal(t, "c", v)
	})
}

func TestLastOrError(t *testing.T) {
	t.Run("Element", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v, err := c.LastOrError()

		assert.Nil(t, err)
		assert.Equal(t, "c", v)
	})

	t.Run("NoElement_Error", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})
		_, err := c.LastOrError()

		assert.NotNil(t, err)
		assert.Equal(t, collection.ErrNoElement, err)
	})
}

func TestCount(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	v := c.Count()

	assert.Equal(t, 3, v)
}

func TestContains(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	t.Run("True", func(t *testing.T) {
		v := c.Contains(func(x string) bool {
			return x == "a"
		})

		assert.True(t, v)
	})

	t.Run("False", func(t *testing.T) {
		v := c.Contains(func(x string) bool {
			return x == "z"
		})

		assert.False(t, v)
	})
}

func TestDistinct(t *testing.T) {
	t.Run("StringsWithDuplicates", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "a", "c", "b"})
		result := c.Distinct(func(a, b string) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, "a")
		assert.Contains(t, result, "b")
		assert.Contains(t, result, "c")
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})
		result := c.Distinct(func(a, b string) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("CustomEquality", func(t *testing.T) {
		type person struct {
			ID   int
			Name string
		}

		c := collection.NewFromSlice([]person{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
			{ID: 3, Name: "alice"}, // Different case, same name
		})

		result := c.Distinct(func(a, b person) bool {
			return strings.EqualFold(a.Name, b.Name)
		}).Slice()

		assert.Equal(t, 2, len(result))
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "a", "c", "b"})
		for range *c.Distinct(func(a, b string) bool {
			return a == b
		}) {
			break
		}
	})
}

func TestSkip(t *testing.T) {
	t.Run("SkipSome", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		result := c.Skip(2).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "c", result[0])
		assert.Equal(t, "d", result[1])
		assert.Equal(t, "e", result[2])
	})

	t.Run("SkipAll", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.Skip(3).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("SkipMore", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.Skip(5).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("SkipZero", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.Skip(0).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
		assert.Equal(t, "c", result[2])
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		for range *c.Skip(2) {
			break
		}
	})
}

func TestSkipUntil(t *testing.T) {
	t.Run("SkipUntilSome", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		result := c.SkipUntil(func(x string) bool {
			return x == "d"
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, "d", result[0])
		assert.Equal(t, "e", result[1])
	})

	t.Run("SkipUntilAll", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.SkipUntil(func(x string) bool {
			return x == "z"
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("SkipUntilNone", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.SkipUntil(func(x string) bool {
			return x == "a"
		}).Slice()

		assert.Equal(t, 3, len(result))
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		for range *c.SkipUntil(func(x string) bool {
			return x == "d"
		}) {
			break
		}
	})
}

func TestTake(t *testing.T) {
	t.Run("TakeSome", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		result := c.Take(3).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
		assert.Equal(t, "c", result[2])
	})

	t.Run("TakeAll", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.Take(3).Slice()

		assert.Equal(t, 3, len(result))
	})

	t.Run("TakeMore", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.Take(5).Slice()

		assert.Equal(t, 3, len(result))
	})

	t.Run("TakeZero", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.Take(0).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		for range *c.Take(3) {
			break
		}
	})
}

func TestTakeUntil(t *testing.T) {
	t.Run("TakeUntilSome", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		result := c.TakeUntil(func(x string) bool {
			return x == "d"
		}).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
		assert.Equal(t, "c", result[2])
	})

	t.Run("TakeUntilAll", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.TakeUntil(func(x string) bool {
			return x == "z"
		}).Slice()

		assert.Equal(t, 3, len(result))
	})

	t.Run("TakeUntilNone", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		result := c.TakeUntil(func(x string) bool {
			return x == "a"
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e"})
		for range *c.TakeUntil(func(x string) bool {
			return x == "d"
		}) {
			break
		}
	})
}

func TestTakeWhile(t *testing.T) {
	t.Run("TakeWhileSome", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		result := c.TakeWhile(func(x int) bool {
			return x < 3
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, 1, result[0])
		assert.Equal(t, 2, result[1])
	})

	t.Run("TakeWhileAll", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		result := c.TakeWhile(func(x int) bool {
			return x < 5
		}).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, 1, result[0])
		assert.Equal(t, 2, result[1])
		assert.Equal(t, 3, result[2])
		assert.Equal(t, 4, result[3])
	})

	t.Run("TakeWhileNone", func(t *testing.T) {
		c := collection.NewFromSlice([]int{})
		result := c.TakeWhile(func(x int) bool {
			return x < 5
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		for range *c.TakeWhile(func(x int) bool {
			return x < 5
		}) {
			break
		}
	})
}

func TestSkipWhile(t *testing.T) {
	t.Run("SkipWhileSome", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		result := c.SkipWhile(func(x int) bool {
			return x < 3
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, 3, result[0])
		assert.Equal(t, 4, result[1])
	})

	t.Run("SkipWhileAll", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		result := c.SkipWhile(func(x int) bool {
			return x < 1
		}).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, 1, result[0])
		assert.Equal(t, 2, result[1])
		assert.Equal(t, 3, result[2])

	})

	t.Run("SkipWhileNone", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		result := c.SkipWhile(func(x int) bool {
			return x < 5
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("Break", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4})
		for range *c.SkipWhile(func(x int) bool {
			return x < 5
		}) {
			break
		}
	})
}

func TestAny(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v := c.Any(func(x string) bool {
			return x == "b"
		})
		assert.True(t, v)
	})

	t.Run("False", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})
		v := c.Any(func(x string) bool {
			return x == "d"
		})
		assert.False(t, v)
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})
		v := c.Any(func(x string) bool {
			return true
		})
		assert.False(t, v)
	})
}

func TestOrderBy(t *testing.T) {
	t.Run("IntAscending", func(t *testing.T) {
		c := collection.NewFromSlice([]int{3, 1, 4, 2})
		result := c.OrderBy(func(x int) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, 1, result[0])
		assert.Equal(t, 2, result[1])
		assert.Equal(t, 3, result[2])
		assert.Equal(t, 4, result[3])
	})

	t.Run("IntDescending", func(t *testing.T) {
		c := collection.NewFromSlice([]int{3, 1, 4, 2})
		result := c.OrderBy(func(x int) any {
			return x
		}, false).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, 4, result[0])
		assert.Equal(t, 3, result[1])
		assert.Equal(t, 2, result[2])
		assert.Equal(t, 1, result[3])
	})

	t.Run("Int8Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]int8{3, 1, 4, 2})
		result := c.OrderBy(func(x int8) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, int8(1), result[0])
		assert.Equal(t, int8(2), result[1])
		assert.Equal(t, int8(3), result[2])
		assert.Equal(t, int8(4), result[3])
	})

	t.Run("Int16Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]int16{3, 1, 4, 2})
		result := c.OrderBy(func(x int16) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, int16(1), result[0])
		assert.Equal(t, int16(2), result[1])
		assert.Equal(t, int16(3), result[2])
		assert.Equal(t, int16(4), result[3])
	})

	t.Run("Int32Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]int32{3, 1, 4, 2})
		result := c.OrderBy(func(x int32) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, int32(1), result[0])
		assert.Equal(t, int32(2), result[1])
		assert.Equal(t, int32(3), result[2])
		assert.Equal(t, int32(4), result[3])
	})

	t.Run("Int64Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]int64{3, 1, 4, 2})
		result := c.OrderBy(func(x int64) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, int64(1), result[0])
		assert.Equal(t, int64(2), result[1])
		assert.Equal(t, int64(3), result[2])
		assert.Equal(t, int64(4), result[3])
	})

	t.Run("UintAscending", func(t *testing.T) {
		c := collection.NewFromSlice([]uint{3, 1, 4, 2})
		result := c.OrderBy(func(x uint) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, uint(1), result[0])
		assert.Equal(t, uint(2), result[1])
		assert.Equal(t, uint(3), result[2])
		assert.Equal(t, uint(4), result[3])
	})

	t.Run("Uint8Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]uint8{3, 1, 4, 2})
		result := c.OrderBy(func(x uint8) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, uint8(1), result[0])
		assert.Equal(t, uint8(2), result[1])
		assert.Equal(t, uint8(3), result[2])
		assert.Equal(t, uint8(4), result[3])
	})

	t.Run("Uint16Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]uint16{3, 1, 4, 2})
		result := c.OrderBy(func(x uint16) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, uint16(1), result[0])
		assert.Equal(t, uint16(2), result[1])
		assert.Equal(t, uint16(3), result[2])
		assert.Equal(t, uint16(4), result[3])
	})

	t.Run("Uint32Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]uint32{3, 1, 4, 2})
		result := c.OrderBy(func(x uint32) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, uint32(1), result[0])
		assert.Equal(t, uint32(2), result[1])
		assert.Equal(t, uint32(3), result[2])
		assert.Equal(t, uint32(4), result[3])
	})

	t.Run("Uint64Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]uint64{3, 1, 4, 2})
		result := c.OrderBy(func(x uint64) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, uint64(1), result[0])
		assert.Equal(t, uint64(2), result[1])
		assert.Equal(t, uint64(3), result[2])
		assert.Equal(t, uint64(4), result[3])
	})

	t.Run("Float32Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]float32{3, 1, 4, 2})
		result := c.OrderBy(func(x float32) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, float32(1), result[0])
		assert.Equal(t, float32(2), result[1])
		assert.Equal(t, float32(3), result[2])
		assert.Equal(t, float32(4), result[3])
	})

	t.Run("Float64Ascending", func(t *testing.T) {
		c := collection.NewFromSlice([]float64{3, 1, 4, 2})
		result := c.OrderBy(func(x float64) any {
			return x
		}, true).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, float64(1), result[0])
		assert.Equal(t, float64(2), result[1])
		assert.Equal(t, float64(3), result[2])
		assert.Equal(t, float64(4), result[3])
	})

	t.Run("StringsAscending", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"banana", "apple", "cherry"})
		result := c.OrderBy(func(x string) any {
			return x
		}, true).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "apple", result[0])
		assert.Equal(t, "banana", result[1])
		assert.Equal(t, "cherry", result[2])
	})

	t.Run("StringsDescending", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"banana", "apple", "cherry"})
		result := c.OrderBy(func(x string) any {
			return x
		}, false).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "cherry", result[0])
		assert.Equal(t, "banana", result[1])
		assert.Equal(t, "apple", result[2])
	})

	t.Run("StructDefault", func(t *testing.T) {
		type test struct{}

		c := collection.NewFromSlice([]test{
			{},
			{},
			{},
		})

		result := c.OrderBy(func(x test) any {
			return x
		}, true).Slice()

		assert.Equal(t, 3, len(result))
	})

	t.Run("StructField", func(t *testing.T) {
		type person struct {
			Name string
			Age  int
		}

		c := collection.NewFromSlice([]person{
			{Name: "Bob", Age: 30},
			{Name: "Alice", Age: 25},
			{Name: "Charlie", Age: 35},
		})

		result := c.OrderBy(func(x person) any {
			return x.Age
		}, true).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "Alice", result[0].Name)
		assert.Equal(t, "Bob", result[1].Name)
		assert.Equal(t, "Charlie", result[2].Name)
	})
}

func TestConcat(t *testing.T) {
	t.Run("BothHaveElements", func(t *testing.T) {
		c1 := collection.NewFromSlice([]string{"a", "b"})
		c2 := collection.NewFromSlice([]string{"c", "d"})

		result := c1.Concat(c2).Slice()

		assert.Equal(t, 4, len(result))
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
		assert.Equal(t, "c", result[2])
		assert.Equal(t, "d", result[3])
	})

	t.Run("FirstEmpty", func(t *testing.T) {
		c1 := collection.NewFromSlice([]string{})
		c2 := collection.NewFromSlice([]string{"c", "d"})

		result := c1.Concat(c2).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, "c", result[0])
		assert.Equal(t, "d", result[1])
	})

	t.Run("SecondEmpty", func(t *testing.T) {
		c1 := collection.NewFromSlice([]string{"a", "b"})
		c2 := collection.NewFromSlice([]string{})

		result := c1.Concat(c2).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, "a", result[0])
		assert.Equal(t, "b", result[1])
	})

	t.Run("BothEmpty", func(t *testing.T) {
		c1 := collection.NewFromSlice([]string{})
		c2 := collection.NewFromSlice([]string{})

		result := c1.Concat(c2).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("BreakFirst", func(t *testing.T) {
		c1 := collection.NewFromSlice([]string{"a", "b"})
		c2 := collection.NewFromSlice([]string{})

		for range *c1.Concat(c2) {
			break
		}
	})
}

func TestGroupBy(t *testing.T) {
	t.Run("SimpleGrouping", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"apple", "banana", "cherry", "apricot", "blueberry"})

		groups := c.GroupBy(func(x string) any {
			return string(x[0]) // Group by first character
		})

		assert.Equal(t, 3, len(groups))

		aGroup := groups["a"].Slice()
		assert.Equal(t, 2, len(aGroup))
		assert.Contains(t, aGroup, "apple")
		assert.Contains(t, aGroup, "apricot")

		bGroup := groups["b"].Slice()
		assert.Equal(t, 2, len(bGroup))
		assert.Contains(t, bGroup, "banana")
		assert.Contains(t, bGroup, "blueberry")

		cGroup := groups["c"].Slice()
		assert.Equal(t, 1, len(cGroup))
		assert.Equal(t, "cherry", cGroup[0])
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		c := collection.NewFromSlice([]string{})

		groups := c.GroupBy(func(x string) any {
			return string(x[0])
		})

		assert.Equal(t, 0, len(groups))
	})

	t.Run("GroupByLength", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "bb", "ccc", "dd", "eee", "f"})

		groups := c.GroupBy(func(x string) any {
			return len(x)
		})

		assert.Equal(t, 3, len(groups))

		group1 := groups[1].Slice()
		assert.Equal(t, 2, len(group1))
		assert.Contains(t, group1, "a")
		assert.Contains(t, group1, "f")

		group2 := groups[2].Slice()
		assert.Equal(t, 2, len(group2))
		assert.Contains(t, group2, "bb")
		assert.Contains(t, group2, "dd")

		group3 := groups[3].Slice()
		assert.Equal(t, 2, len(group3))
		assert.Contains(t, group3, "ccc")
		assert.Contains(t, group3, "eee")
	})
}

func TestUnion(t *testing.T) {
	t.Run("WithDuplicates", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2, 3, 4})
		c2 := collection.NewFromSlice([]int{3, 4, 5, 6})

		result := c1.Union(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 6, len(result))
		assert.Contains(t, result, 1)
		assert.Contains(t, result, 2)
		assert.Contains(t, result, 3)
		assert.Contains(t, result, 4)
		assert.Contains(t, result, 5)
		assert.Contains(t, result, 6)
	})

	t.Run("FirstEmpty", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{})
		c2 := collection.NewFromSlice([]int{3, 4, 5})

		result := c1.Union(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, 3)
		assert.Contains(t, result, 4)
		assert.Contains(t, result, 5)
	})

	t.Run("SecondEmpty", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2, 3})
		c2 := collection.NewFromSlice([]int{})

		result := c1.Union(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, 1)
		assert.Contains(t, result, 2)
		assert.Contains(t, result, 3)
	})
}

func TestIntersect(t *testing.T) {
	t.Run("WithCommonElements", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2, 3, 4})
		c2 := collection.NewFromSlice([]int{3, 4, 5, 6})

		result := c1.Intersect(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Contains(t, result, 3)
		assert.Contains(t, result, 4)
	})

	t.Run("NoCommonElements", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2})
		c2 := collection.NewFromSlice([]int{3, 4})

		result := c1.Intersect(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("EmptyFirst", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{})
		c2 := collection.NewFromSlice([]int{3, 4})

		result := c1.Intersect(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("EmptySecond", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2})
		c2 := collection.NewFromSlice([]int{})

		result := c1.Intersect(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 0, len(result))
	})
}

func TestExcept(t *testing.T) {
	t.Run("WithCommonElements", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2, 3, 4})
		c2 := collection.NewFromSlice([]int{3, 4, 5, 6})

		result := c1.Except(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Contains(t, result, 1)
		assert.Contains(t, result, 2)
	})

	t.Run("NoCommonElements", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2})
		c2 := collection.NewFromSlice([]int{3, 4})

		result := c1.Except(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Contains(t, result, 1)
		assert.Contains(t, result, 2)
	})

	t.Run("AllElementsCommon", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2})
		c2 := collection.NewFromSlice([]int{1, 2, 3})

		result := c1.Except(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 0, len(result))
	})

	t.Run("EmptyFirst", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{})
		c2 := collection.NewFromSlice([]int{3, 4})

		result := c1.Except(c2, func(a, b int) bool {
			return a == b
		}).Slice()

		assert.Equal(t, 0, len(result))
	})
}

func TestReverse(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	result := c.Reverse().Slice()

	assert.Equal(t, 3, len(result))
	assert.Equal(t, "c", result[0])
	assert.Equal(t, "b", result[1])
	assert.Equal(t, "a", result[2])
}

func TestAppend(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	result := c.Append("d").Slice()

	assert.Equal(t, 4, len(result))
	assert.Equal(t, "a", result[0])
	assert.Equal(t, "b", result[1])
	assert.Equal(t, "c", result[2])
	assert.Equal(t, "d", result[3])
}

func TestPrepend(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	result := c.Prepend("d").Slice()

	assert.Equal(t, 4, len(result))
	assert.Equal(t, "d", result[0])
	assert.Equal(t, "a", result[1])
	assert.Equal(t, "b", result[2])
	assert.Equal(t, "c", result[3])
}

func TestChunk(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c", "d", "e", "f", "g", "h"})
	result := c.Chunk(3)

	assert.Equal(t, 3, len(result))

	assert.Equal(t, 3, len(result[0].Slice()))
	assert.Equal(t, "a", result[0].Slice()[0])
	assert.Equal(t, "b", result[0].Slice()[1])
	assert.Equal(t, "c", result[0].Slice()[2])

	assert.Equal(t, 3, len(result[1].Slice()))
	assert.Equal(t, "d", result[1].Slice()[0])
	assert.Equal(t, "e", result[1].Slice()[1])
	assert.Equal(t, "f", result[1].Slice()[2])

	assert.Equal(t, 2, len(result[2].Slice()))
	assert.Equal(t, "g", result[2].Slice()[0])
	assert.Equal(t, "h", result[2].Slice()[1])
}

func TestAggregate(t *testing.T) {
	t.Run("Logic", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"Amsterdam", "Berlin", "New York", "San Francisco"})

		result := c.Aggregate("Paris", func(accumulator any, item string) any {
			if len(accumulator.(string)) < len(item) {
				return item
			}

			return accumulator
		})

		assert.Equal(t, "San Francisco", result)
	})

	t.Run("Sum", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3, 4, 5})

		result := c.Aggregate(0, func(accumulator any, item int) any {
			return accumulator.(int) + item
		})

		assert.Equal(t, 15, result)
	})

	t.Run("Concatenation", func(t *testing.T) {
		c := collection.NewFromSlice([]string{"a", "b", "c"})

		result := c.Aggregate("", func(accumulator any, item string) any {
			return accumulator.(string) + item
		})

		assert.Equal(t, "abc", result)
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		c := collection.NewFromSlice([]int{})

		seed := 10
		result := c.Aggregate(seed, func(accumulator any, item int) any {
			assert.Fail(t, "This should not be called")
			return nil
		})

		assert.Equal(t, seed, result)
	})
}

func TestZip(t *testing.T) {
	t.Run("EqualLength", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2, 3})
		c2 := collection.NewFromSlice([]string{"a", "b", "c"})

		result := collection.Zip(c1, c2, func(a int, b string) string {
			return strconv.Itoa(a) + b
		}).Slice()

		assert.Equal(t, 3, len(result))
		assert.Equal(t, "1a", result[0])
		assert.Equal(t, "2b", result[1])
		assert.Equal(t, "3c", result[2])
	})

	t.Run("FirstShorter", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2})
		c2 := collection.NewFromSlice([]string{"a", "b", "c"})

		result := collection.Zip(c1, c2, func(a int, b string) string {
			return strconv.Itoa(a) + b
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, "1a", result[0])
		assert.Equal(t, "2b", result[1])
	})

	t.Run("SecondShorter", func(t *testing.T) {
		c1 := collection.NewFromSlice([]int{1, 2, 3})
		c2 := collection.NewFromSlice([]string{"a", "b"})

		result := collection.Zip(c1, c2, func(a int, b string) string {
			return strconv.Itoa(a) + b
		}).Slice()

		assert.Equal(t, 2, len(result))
		assert.Equal(t, "1a", result[0])
		assert.Equal(t, "2b", result[1])
	})
}

func TestSlice(t *testing.T) {
	c := collection.NewFromSlice([]string{"a", "b", "c"})
	v := c.Slice()

	assert.Equal(t, []string{"a", "b", "c"}, v)
}

func TestAverage(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		c := collection.NewFromSlice([]uint{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Uint8", func(t *testing.T) {
		c := collection.NewFromSlice([]uint8{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Uint16", func(t *testing.T) {
		c := collection.NewFromSlice([]uint16{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Uint32", func(t *testing.T) {
		c := collection.NewFromSlice([]uint32{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Uint64", func(t *testing.T) {
		c := collection.NewFromSlice([]uint64{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Int", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Int8", func(t *testing.T) {
		c := collection.NewFromSlice([]int8{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Int16", func(t *testing.T) {
		c := collection.NewFromSlice([]int16{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Int32", func(t *testing.T) {
		c := collection.NewFromSlice([]int32{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Int64", func(t *testing.T) {
		c := collection.NewFromSlice([]int64{1, 2, 3})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Float32", func(t *testing.T) {
		c := collection.NewFromSlice([]float32{1.0, 2.0, 3.0})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})

	t.Run("Float64", func(t *testing.T) {
		c := collection.NewFromSlice([]float64{1.0, 2.0, 3.0})
		v := collection.Average(c)
		f, _ := v.Float64()

		assert.Equal(t, 2.0, f)
	})
}

func TestSum(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		c := collection.NewFromSlice([]uint{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Uint8", func(t *testing.T) {
		c := collection.NewFromSlice([]uint8{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Uint16", func(t *testing.T) {
		c := collection.NewFromSlice([]uint16{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Uint32", func(t *testing.T) {
		c := collection.NewFromSlice([]uint32{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Uint64", func(t *testing.T) {
		c := collection.NewFromSlice([]uint64{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Int", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Int8", func(t *testing.T) {
		c := collection.NewFromSlice([]int8{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Int16", func(t *testing.T) {
		c := collection.NewFromSlice([]int16{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Int32", func(t *testing.T) {
		c := collection.NewFromSlice([]int32{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Int64", func(t *testing.T) {
		c := collection.NewFromSlice([]int64{1, 2, 3})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Float32", func(t *testing.T) {
		c := collection.NewFromSlice([]float32{1.0, 2.0, 3.0})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})

	t.Run("Float64", func(t *testing.T) {
		c := collection.NewFromSlice([]float64{1.0, 2.0, 3.0})
		v := collection.Sum(c)
		f, _ := v.Float64()

		assert.Equal(t, 6.0, f)
	})
}

func TestSumInt(t *testing.T) {
	t.Run("Uint", func(t *testing.T) {
		c := collection.NewFromSlice([]uint{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Uint8", func(t *testing.T) {
		c := collection.NewFromSlice([]uint8{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Uint16", func(t *testing.T) {
		c := collection.NewFromSlice([]uint16{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Uint32", func(t *testing.T) {
		c := collection.NewFromSlice([]uint32{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Int", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Int8", func(t *testing.T) {
		c := collection.NewFromSlice([]int8{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Int16", func(t *testing.T) {
		c := collection.NewFromSlice([]int16{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})

	t.Run("Int32", func(t *testing.T) {
		c := collection.NewFromSlice([]int32{1, 2, 3})
		v := collection.SumInt(c)
		f := v.Int64()

		assert.Equal(t, int64(6), f)
	})
}

func TestMin(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {

		c := collection.NewFromSlice([]int{1, 2, 3})
		v := collection.Min(c)

		assert.Equal(t, 1, v)
	})

	t.Run("Negative", func(t *testing.T) {

		c := collection.NewFromSlice([]int{-1, -2, -3})
		v := collection.Min(c)

		assert.Equal(t, -3, v)
	})
}

func TestMax(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		c := collection.NewFromSlice([]int{1, 2, 3})
		v := collection.Max(c)

		assert.Equal(t, 3, v)
	})

	t.Run("Negative", func(t *testing.T) {
		c := collection.NewFromSlice([]int{-1, -2, -3})
		v := collection.Max(c)

		assert.Equal(t, -1, v)
	})
}
