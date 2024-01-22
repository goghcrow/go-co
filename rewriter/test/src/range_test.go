package src

import (
	"sort"
	"testing"

	. "github.com/goghcrow/go-co"
)

func TestRangeString(t *testing.T) {
	{
		g := func() (_ Iter[int]) {
			for range "" {
				Yield(1)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int(nil))
	}

	{
		g := func() (_ Iter[int]) {
			for range "hello" {
				Yield(1)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{1, 1, 1, 1, 1})
	}

	{
		g := func() (_ Iter[int]) {
			for k := range "hello" {
				Yield(k)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{0, 1, 2, 3, 4})
	}

	{
		g := func() (_ Iter[int]) {
			for k, _ := range "hello" {
				Yield(k)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{0, 1, 2, 3, 4})
	}

	{
		g := func() (_ Iter[rune]) {
			for _, v := range "hello" {
				Yield(v)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []rune("hello"))
	}
}

func TestRangeSlice(t *testing.T) {
	{
		g := func() (_ Iter[int]) {
			var xs []int
			for range xs {
				Yield(1)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int(nil))
	}

	xs := []int{1, 2, 3, 4, 5}

	{
		g := func() (_ Iter[int]) {
			for range xs {
				Yield(1)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{1, 1, 1, 1, 1})
	}

	{
		g := func() (_ Iter[int]) {
			for k := range xs {
				Yield(k)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{0, 1, 2, 3, 4})
	}

	{
		g := func() (_ Iter[int]) {
			for k, _ := range xs {
				Yield(k)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{0, 1, 2, 3, 4})
	}

	{
		g := func() (_ Iter[int]) {
			for _, v := range xs {
				Yield(v)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{1, 2, 3, 4, 5})
	}
}

// func TestRangeArray(t *testing.T) {
// 	{
// 		g := func() (_ Iter[int]) {
// 			var xs [0]int
// 			for range xs {
// 				Yield(1)
// 			}
// 			return
// 		}
// 		assertEqual(t, iter2slice(g()), []int(nil))
// 	}
//
// 	xs := [5]int{1, 2, 3, 4, 5}
//
// 	{
// 		g := func() (_ Iter[int]) {
// 			for range xs {
// 				Yield(1)
// 			}
// 			return
// 		}
// 		assertEqual(t, iter2slice(g()), []int{1, 1, 1, 1, 1})
// 	}
//
// 	{
// 		g := func() (_ Iter[int]) {
// 			for k := range xs {
// 				Yield(k)
// 			}
// 			return
// 		}
// 		assertEqual(t, iter2slice(g()), []int{0, 1, 2, 3, 4})
// 	}
//
// 	{
// 		g := func() (_ Iter[int]) {
// 			for k, _ := range xs {
// 				Yield(k)
// 			}
// 			return
// 		}
// 		assertEqual(t, iter2slice(g()), []int{0, 1, 2, 3, 4})
// 	}
//
// 	{
// 		g := func() (_ Iter[int]) {
// 			for _, v := range xs {
// 				Yield(v)
// 			}
// 			return
// 		}
// 		assertEqual(t, iter2slice(g()), []int{1, 2, 3, 4, 5})
// 	}
// }

func TestRangeMap(t *testing.T) {
	{
		g := func() (_ Iter[int]) {
			var xs map[string]int
			for range xs {
				Yield(1)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int(nil))
	}

	xs := map[string]int{
		"1": 1, "2": 2, "3": 3, "4": 4, "5": 5,
	}

	{
		g := func() (_ Iter[int]) {
			for range xs {
				Yield(1)
			}
			return
		}
		assertEqual(t, iter2slice(g()), []int{1, 1, 1, 1, 1})
	}

	{
		g := func() (_ Iter[string]) {
			for k := range xs {
				Yield(k)
			}
			return
		}

		s := iter2slice(g())
		sort.Strings(s)
		assertEqual(t, s, []string{"1", "2", "3", "4", "5"})
	}

	{
		g := func() (_ Iter[string]) {
			for k, _ := range xs {
				Yield(k)
			}
			return
		}

		s := iter2slice(g())
		sort.Strings(s)
		assertEqual(t, s, []string{"1", "2", "3", "4", "5"})
	}

	{
		g := func() (_ Iter[int]) {
			for _, v := range xs {
				Yield(v)
			}
			return
		}

		s := iter2slice(g())
		sort.Ints(s)
		assertEqual(t, s, []int{1, 2, 3, 4, 5})
	}
}
