package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

func trivalSwitch() (_ Iter[int]) {
	switch a := 1; a {
	case 1:
		return
	case 2:
		return
	default:
		return
	}
}

func TestSwitchInitNameConflict(t *testing.T) {
	g := func() (_ Iter[int]) {
		a := 42
		switch a := a + 1; a {
		case 1:
			assertEqual(t, a, 43)
			return
		case 2:
			assertEqual(t, a, 43)
			return
		default:
			assertEqual(t, a, 43)
			return
		}
	}
	iter2slice(g())
}

func TestSwitchWithYieldInInit(t *testing.T) {
	g := func() (_ Iter[int]) {
		switch Yield(1); 1 {
		case 1:
			return
		case 2:
			return
		default:
			return
		}
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{1})
}

func TestSwitchWithYieldInInitAndCase(t *testing.T) {
	g := func(a int) (_ Iter[int]) {
		switch Yield(a); a {
		case 1:
			Yield(1)
		case 2:
			Yield(2)
		default:
			Yield(42)
		}
		return
	}
	{
		xs := iter2slice(g(1))
		assertEqual(t, xs, []int{1, 1})
	}
	{
		xs := iter2slice(g(2))
		assertEqual(t, xs, []int{2, 2})
	}
	{
		xs := iter2slice(g(3))
		assertEqual(t, xs, []int{3, 42})
	}
}

func TestSwitchWithYieldInInitAndCaseWithoutDefault(t *testing.T) {
	g := func(a int) (_ Iter[int]) {
		switch Yield(a); a {
		case 1:
			Yield(1)
		case 2:
			Yield(2)
		}
		return
	}
	{
		xs := iter2slice(g(1))
		assertEqual(t, xs, []int{1, 1})
	}
	{
		xs := iter2slice(g(2))
		assertEqual(t, xs, []int{2, 2})
	}
	{
		xs := iter2slice(g(3))
		assertEqual(t, xs, []int{3})
	}
}

func TestSwitch(t *testing.T) {
	i := 0
	f := func() { i++ }
	g := func(a, b int) (_ Iter[string]) {
		switch i = 0; a {
		case 1:
			switch b {
			case 1:
				Yield("11")
			case 2:
				Yield("12")
			default:
				Yield("1?")
			}
			f()
		case 2:
			switch b {
			case 1:
				Yield("21")
			case 2:
				Yield("22")
			default:
				Yield("2?")
			}
			f()
		}
		f()
		if i == 1 {
			Yield("f")
		} else if i == 2 {
			Yield("ff")
		} else {
			panic("unexpected")
		}
		return
	}

	assertEqual(t, iter2slice(g(1, 1)), []string{"11", "ff"})
	assertEqual(t, iter2slice(g(1, 2)), []string{"12", "ff"})
	assertEqual(t, iter2slice(g(1, 3)), []string{"1?", "ff"})
	assertEqual(t, iter2slice(g(2, 1)), []string{"21", "ff"})
	assertEqual(t, iter2slice(g(2, 2)), []string{"22", "ff"})
	assertEqual(t, iter2slice(g(2, 3)), []string{"2?", "ff"})
	assertEqual(t, iter2slice(g(3, 1)), []string{"f"})
}
