package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

func TestBlockStmt(t *testing.T) {
	g := func() Iter[int] {
		{
			Yield(1)
		}
		Yield(2)
		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{1, 2})
}

func TestTrivalBlock(t *testing.T) {
	g := func() Iter[int] {
		for {
			{
				if true {
					break
				} else {
					continue
				}
			}
		}
		Yield(1)
		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{1})
}

func TestYieldBlock(t *testing.T) {
	g := func() Iter[int] {
		for i := 0; i < 5; i++ {
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
	assertEqual(t, xs, []int{0, 1, 2, 3, 4})
}

func TestYieldBlock2(t *testing.T) {
	f := func() {}
	g := func() Iter[int] {
		for i := 0; i < 5; i++ {
			{
				if true {
					Yield(i)
				} else {
					break
				}
			}
			Yield(1)
			f()
		}
		return nil
	}
	xs := iter2slice(g())
	assertEqual(t, xs, []int{0, 1, 1, 1, 2, 1, 3, 1, 4, 1})
}
