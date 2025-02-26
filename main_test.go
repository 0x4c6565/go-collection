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

// println(c.Where(func(x string) bool {
// 	return x == "b"
// }).First())

// println(c.Last())

// var somestructs []somestruct
// somestructs = append(somestructs, somestruct{
// 	Property1: "s1",
// 	Property2: 1,
// }, somestruct{
// 	Property1: "s2",
// 	Property2: 2,
// })

// s := golinq.NewFromSlice(somestructs)
// for _, val := range s.Select(func(x somestruct) any {
// 	return x.Property1
// }).Slice() {
// 	println(val.(string))
// }
