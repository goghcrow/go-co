package seq

import (
	"testing"
)

func TestMapIter(t *testing.T) {
	m := map[string]int{
		"hello": 42,
		"world": 100,
	}
	iter := NewMapIter(m)
	for iter.MoveNext() {
		current := iter.Current()
		t.Log(current.Key, current.Val)
	}
}

func Yield123() Iterator[int] {
	// yield 1
	// yield 2
	// yield 3
	return Start(Delay(func() Seq[int] {
		return Bind(1, func() Seq[int] {
			return Bind(2, func() Seq[int] {
				return Bind(3, func() Seq[int] {
					return Normal[int]()
				})
			})
		})
	}))
}

func TestYield123(t *testing.T) {
	iter := Yield123()
	for iter.MoveNext() {
		println(iter.Current())
	}
}

func YieldEven() Iterator[int] {
	//	for i := 0; i < 10; i++ {
	//	 	if i % 2 == 0 {
	//			 continue
	//		}
	//		yield i
	//	}
	return Start(Delay(func() Seq[int] {
		i := 0
		return For(
			func() bool { return i < 10 },
			func() { i++ },
			Combine(Delay[int](func() Seq[int] {
				if i%2 == 0 {
					return Continue[int]()
				}
				return Normal[int]()
			}), Delay[int](func() Seq[int] {
				return Bind[int](i, func() Seq[int] {
					return Normal[int]()
				})
			})),
		)
	}))
}

func TestForYieldEven(t *testing.T) {
	iter := YieldEven()
	for iter.MoveNext() {
		println(iter.Current())
	}
}
