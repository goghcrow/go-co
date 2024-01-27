package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

func TestForRange(t *testing.T) {
	yield123 := func() Iter[int] {
		Yield(1)
		Yield(2)
		Yield(3)
		return nil
	}

	var xs []int
	for v := range yield123() {
		xs = append(xs, v)
	}
	assertEqual(t, xs, []int{1, 2, 3})

	var ys []int
	iter := yield123()
	for iter.MoveNext() {
		ys = append(ys, iter.Current())
		ys = append(ys, iter.Current())
	}
	assertEqual(t, ys, []int{1, 1, 2, 2, 3, 3})
}

func Test123(t *testing.T) {
	yield123 := func() Iter[int] {
		Yield(1)
		Yield(2)
		Yield(3)
		return nil
	}
	xs := iter2slice(yield123())
	assertEqual(t, xs, []int{1, 2, 3})

	{
		assertEqual(t, iter2slice(func() Iter[int] {
			Yield(1)
			Yield(2)
			Yield(3)
			return nil
		}()), []int{1, 2, 3})
	}
}

func TestYieldFrom(t *testing.T) {
	var (
		from = func() Iter[int] {
			Yield(1)
			Yield(2)
			return nil
		}
		gen = func() Iter[int] {
			YieldFrom(from())
			Yield(3)
			return nil
		}
	)
	xs := iter2slice(gen())
	assertEqual(t, xs, []int{1, 2, 3})
}

func TestYieldABC(t *testing.T) {
	f := func() {}
	g := func() Iter[int] {
		Yield(1)
		f()

		if false {
			Yield(2)
			f()
		} else {
			for i := 0; i < 3; i++ {
				Yield(i)
				f()
			}
		}

		YieldFrom(func() Iter[int] {
			Yield(1)
			return nil
		}())

		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{1, 0, 1, 2, 1})
}

func TestYieldFunc(t *testing.T) {
	gen := func() Iter[func() string] {
		Yield(func() string { return "hello" })
		Yield(func() string { return "world" })
		return nil
	}
	var xs []string
	for f := range gen() {
		xs = append(xs, f())
	}
	assertEqual(t, xs, []string{"hello", "world"})
}

func TestRecursive1(t *testing.T) {
	recGen := func() Iter[int] {
		var rec func(int) Iter[int]
		rec = func(n int) (_ Iter[int]) {
			if n == 0 {
				return
			}
			Yield(0)
			YieldFrom(rec(n - 1))
			return
		}

		YieldFrom(rec(5))
		return nil
	}

	xs := iter2slice(recGen())
	assertEqual(t, xs, []int{0, 0, 0, 0, 0})
}

func TestRecursive2(t *testing.T) {
	var from func(a int) Iter[int]
	from = func(a int) (_ Iter[int]) {
		Yield(1 + a)
		if a <= 3 {
			YieldFrom(from(a + 3))
			YieldFrom(from(a + 6))
		}
		Yield(2 + a)
		return
	}
	gen := func() (_ Iter[int]) {
		YieldFrom(from(0))
		return
	}
	xs := iter2slice(gen())
	assertEqual(t, xs, []int{1, 4, 7, 8, 10, 11, 5, 7, 8, 2})
}

func TestDeepRecursive(t *testing.T) {
	from := func(i int) (_ Iter[int]) {
		Yield(i)
		return
	}
	var gen func(int) Iter[int]
	gen = func(i int) (_ Iter[int]) {
		if i < 50000 {
			YieldFrom(gen(i + 1))
		} else {
			Yield(i)
			YieldFrom(from(i + 1))
		}
		return
	}
	xs := iter2slice(gen(0))
	assertEqual(t, xs, []int{50000, 50001})
}

func TestYieldFromSameGen(t *testing.T) {
	var gen func(int) Iter[int]
	gen = func(a int) (_ Iter[int]) {
		Yield(1 + a)
		if a < 1 {
			YieldFrom(gen(a + 1))
		}
		Yield(3 + a)
		return
	}
	bar := func(gen Iter[int]) (_ Iter[int]) {
		YieldFrom(gen)
		return
	}

	assertEqual(t, iter2slice(bar(gen(0))), []int{1, 2, 4, 3})

	g := gen(0)
	a, b := bar(g), bar(g)

	var xs, ys []int
	for {
		if a.MoveNext() {
			xs = append(xs, a.Current())
		} else {
			break
		}
		if b.MoveNext() {
			ys = append(ys, b.Current())
		} else {
			break
		}
	}
	assertEqual(t, xs, []int{1, 4})
	assertEqual(t, ys, []int{2, 3})
}
