package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

func TestTypeSwitchTrival(t *testing.T) {
	f := func() {}
	g := func() (_ Iter[int]) {
		Yield(0)
		var a any
		switch a.(type) {
		case int:
			Yield(1)
		case string:
			Yield(2)
		default:
			f()
			Yield(3)
			f()
		}
		Yield(4)
		f()
		return
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 3, 4})
}

func TestTypeSwitchTrivalInitAndNotDefaultBranch(t *testing.T) {
	g := func(a any) (_ Iter[int]) {
		Yield(0)
		switch x := a; x.(type) {
		case int:
			Yield(1)
		case string:
			Yield(2)
		default:
			Yield(3)
		}
		Yield(4)
		return
	}

	{
		xs := iter2slice(g(1))
		assertEqual(t, xs, []int{0, 1, 4})
	}
	{
		xs := iter2slice(g(""))
		assertEqual(t, xs, []int{0, 2, 4})
	}
	{
		xs := iter2slice(g(3.13))
		assertEqual(t, xs, []int{0, 3, 4})
	}
}

func TestTypeSwitchScope(t *testing.T) {
	g := func(a any) (_ Iter[int]) {
		Yield(0)
		x := 42
		switch x := a; x.(type) {
		case int:
			Yield(1)
		case string:
			Yield(2)
		default:
			Yield(3)
		}
		Yield(x)
		return
	}

	{
		xs := iter2slice(g(1))
		assertEqual(t, xs, []int{0, 1, 42})
	}
	{
		xs := iter2slice(g(""))
		assertEqual(t, xs, []int{0, 2, 42})
	}
	{
		xs := iter2slice(g(3.13))
		assertEqual(t, xs, []int{0, 3, 42})
	}
}

func TestTypeSwitchYieldInitAndNotDefaultBranch(t *testing.T) {
	g := func() (_ Iter[int]) {
		var a any
		switch Yield(0); a.(type) {
		case int:
			Yield(1)
		case string:
			Yield(2)
		}
		Yield(3)
		return
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 3})
}

func TestTypeSwitchYieldInitAndNotDefaultBranch1(t *testing.T) {
	g := func() (_ Iter[int]) {
		var a any = "hello"
		switch Yield(0); a.(type) {
		case int:
			Yield(1)
		case string:
			Yield(2)
		}
		Yield(3)
		return
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 2, 3})
}

func TestTypeSwitchYieldInitAndNotDefaultBranch2(t *testing.T) {
	g := func() (_ Iter[int]) {
		var a any = "hello"
		switch Yield(0); a.(type) {
		case int:
			Yield(1)
		case string:
			Yield(2)
			return
		}
		Yield(3)
		return
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 2})
}

func TestTypeSwitchYieldInitAndAssign(t *testing.T) {
	g := func() (_ Iter[int]) {
		var a any = 42
		switch Yield(0); b := a.(type) {
		case int:
			Yield(b)
			return
		case string:
			Yield(2)
		}
		Yield(3)
		return
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 42})
}
