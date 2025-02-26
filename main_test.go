package gocollection_test

import (
	"testing"

	gocollection "github.com/0x4c6565/go-collection"
	"github.com/stretchr/testify/assert"
)

func TestFirst(t *testing.T) {
	c := gocollection.NewFromSlice([]string{"a", "b", "c"})
	v := c.First()

	assert.Equal(t, "a", v)
}

func TestLast(t *testing.T) {
	c := gocollection.NewFromSlice([]string{"a", "b", "c"})
	v := c.Last()

	assert.Equal(t, "c", v)
}

func TestCount(t *testing.T) {
	c := gocollection.NewFromSlice([]string{"a", "b", "c"})
	v := c.Count()

	assert.Equal(t, 3, v)
}

func TestContains(t *testing.T) {
	c := gocollection.NewFromSlice([]string{"a", "b", "c"})
	t.Run("true", func(t *testing.T) {
		v := c.Contains(func(x string) bool {
			return x == "a"
		})

		assert.True(t, v)
	})

	t.Run("false", func(t *testing.T) {
		v := c.Contains(func(x string) bool {
			return x == "z"
		})

		assert.False(t, v)
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
	c := gocollection.NewFromSlice(teststructs)

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
}

func TestWhere(t *testing.T) {
	c := gocollection.NewFromSlice([]string{"a", "b", "c"})
	t.Run("true", func(t *testing.T) {
		v := c.Where(func(x string) bool {
			return x == "a"
		}).Slice()

		assert.Len(t, v, 1)
		assert.Equal(t, "a", v[0])
	})
}
