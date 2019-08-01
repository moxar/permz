package permz

import (
	"fmt"
	"testing"
)

func ExamplePermissionResolver() {

	// Given a very simple InArray function
	InArray := func(haystack []int, needle int) bool {
		for _, v := range haystack {
			if v == needle {
				return true
			}
		}
		return false
	}

	// And a set of rights
	var permissions = []int{1, 2, 3}

	// This resolver returns true if the right is granted.
	var Resolve PermissionResolver
	Resolve = func(right Right) bool {
		if v, ok := right.(int); ok {
			return InArray(permissions, v)
		}
		return false
	}

	fmt.Println(Resolve(1))
	fmt.Println(Resolve(7))
	// Output:
	// true
	// false

}

func TestAnd(t *testing.T) {

	type Case struct {
		Sentence string
		A, B     PermissionResolver
		Want     bool
	}

	var cases = []Case{
		{
			Sentence: "true x true",
			A:        True,
			B:        True,
			Want:     true,
		},
		{
			Sentence: "true x false",
			A:        True,
			B:        False,
			Want:     false,
		},
		{
			Sentence: "false x false",
			A:        False,
			B:        False,
			Want:     false,
		},
		{
			Sentence: "true x nil",
			A:        True,
			B:        nil,
			Want:     true,
		},
		{
			Sentence: "false x nil",
			A:        False,
			B:        nil,
			Want:     false,
		},
	}

	for _, c := range cases {
		t.Run(c.Sentence, func(t *testing.T) {
			if And(c.A, c.B)(0) != c.Want {
				t.Error("should be", c.Want)
			}
		})
	}
}

func TestOr(t *testing.T) {

	type Case struct {
		Sentence string
		A, B     PermissionResolver
		Want     bool
	}

	var cases = []Case{
		{
			Sentence: "true x true",
			A:        True,
			B:        True,
			Want:     true,
		},
		{
			Sentence: "true x false",
			A:        True,
			B:        False,
			Want:     true,
		},
		{
			Sentence: "false x false",
			A:        False,
			B:        False,
			Want:     false,
		},
		{
			Sentence: "true x nil",
			A:        True,
			B:        nil,
			Want:     true,
		},
		{
			Sentence: "false x nil",
			A:        False,
			B:        nil,
			Want:     false,
		},
	}

	for _, c := range cases {
		t.Run(c.Sentence, func(t *testing.T) {
			if Or(c.A, c.B)(0) != c.Want {
				t.Error("should be", c.Want)
			}
		})
	}
}
