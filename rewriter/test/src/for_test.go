package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

func TestForPostScope(t *testing.T) {
	g := func() Iter[int] {
		a := 42
		for cnt := 5; cnt > 0; Yield(func() int {
			cnt--
			a++
			return a - 1
		}()) {
			a := 100
			Yield(a)
			a++
		}
		return nil
	}

	xs := iter2slice(g())
	assertEqual(t, xs, []int{100, 42, 100, 43, 100, 44, 100, 45, 100, 46})
}

func TestForBodyScope(t *testing.T) {
	g := func() Iter[int] {
		for i := 0; i < 5; i++ {
			i := 0
			{
				if true {
					Yield(i)
				} else {
					break
				}
			}
		}
		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 0, 0, 0, 0})
}

func TestForBodyScope1(t *testing.T) {
	g := func() Iter[int] {
		for i := 0; i < 5; i++ {
			{
				i := 0
				if true {
					Yield(i)
				} else {
					break
				}
			}
		}
		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 0, 0, 0, 0})
}

func TestFor(t *testing.T) {
	{
		g := func() Iter[int] {
			i := 0
			for Yield(1); i < 3; Yield(2) {
				Yield(3)
				i++
			}
			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{1, 3, 2, 3, 2, 3, 2})
	}

	{
		g := func() Iter[int] {
			i := 0
			Yield(1)
			for i < 3 {
				Yield(3)
				i++
				Yield(2)
			}
			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{1, 3, 2, 3, 2, 3, 2})
	}

	{
		g := func() Iter[int] {
			for Yield(1); false; {
			}
			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{1})
	}

	{
		g := func() Iter[int] {
			flag := true
			for Yield(1); flag; {
				flag = false
			}
			Yield(2)
			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{1, 2})
	}

	{
		g := func() Iter[int] {
			flag := true
			for ; flag; Yield(1) {
				flag = false
			}
			Yield(2)
			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{1, 2})
	}

	{
		g := func() Iter[int] {
			i := 0
			Yield(1)
			for i < 3 {
				Yield(3)
				i++
				Yield(2)
			}
			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{1, 3, 2, 3, 2, 3, 2})
	}

	{
		g := func() Iter[int] {
			i := 0
			Yield(0)
			for ; i < 3; YieldFrom(func() Iter[int] {
				Yield(1)
				Yield(2)
				return nil
			}()) {
				i++
				Yield(42)
			}

			return nil
		}
		xs := iter2slice(g())
		assertEqual(t, xs, []int{0, 42, 1, 2, 42, 1, 2, 42, 1, 2})
	}

	{
		var xs []int
		f := func(g Iter[int]) {
			for i := range g {
				xs = append(xs, i)
			}
		}

		g := func() Iter[int] {
			flag := true
			for ; flag; f(func() Iter[int] {
				Yield(1)
				Yield(2)
				Yield(3)
				return nil
			}()) {
				flag = false
			}
			return nil
		}
		iter2slice(g())
		assertEqual(t, xs, []int{1, 2, 3})

	}
}

func TestReturnBug(t *testing.T) {
	g := func() Iter[int] {
		i := 0
		for ; i < 10; i++ {
			if false {
				break
			}
		}
		Yield(i)
		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{10})
}
