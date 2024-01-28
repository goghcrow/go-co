package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

// 1. Init stmt in ForStmt will up to outer scope
// 2. Key, Value Expr in RangeStmt will be down to body scope

func forInit() Iter[int] {
	i := 0
	for i := 0; i < 3; i++ {
		i := i + 1
		Yield(i)
	}
	Yield(i)
	return nil
}

// ↑ is equivalent to ↓
func forInit1() Iter[int] {
	i := 0
	{
		i := 0
		for ; i < 3; i++ {
			i := i + 1
			Yield(i)
		}
	}
	Yield(i)
	return nil
}

func rangeInit() Iter[int] {
	i := 0
	for _, i := range []int{0, 1, 2} {
		i := i + 1
		Yield(i)
	}
	Yield(i)
	return nil
}

// ↑ is equivalent to ↓
func rangeInit1() Iter[int] {
	i := 0
	for _, i := range []int{0, 1, 2} {
		{
			i := i + 1
			Yield(i)
		}
	}
	Yield(i)
	return nil
}

func TestShadow(t *testing.T) {
	assertEqual(t, iter2slice(forInit()), []int{1, 2, 3, 0})
	assertEqual(t, iter2slice(forInit1()), []int{1, 2, 3, 0})
	assertEqual(t, iter2slice(rangeInit()), []int{1, 2, 3, 0})
	assertEqual(t, iter2slice(rangeInit1()), []int{1, 2, 3, 0})
}
