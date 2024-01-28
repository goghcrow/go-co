package seq

import (
	"reflect"
	"testing"
)

func TestSeq(t *testing.T) {
	all := []struct {
		name    string
		factory func() Iterator[int]
		expect  []int
	}{
		{
			name: "Yield123",
			factory: func() Iterator[int] {
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
			},
			expect: []int{1, 2, 3},
		},
		{
			name: "YieldEven",
			factory: func() Iterator[int] {
				//	for i := 0; i < 10; i++ {
				//	 	if i % 2 == 0 {
				//			 continue
				//		}
				//		yield i
				//	}
				return Start(Delay(func() Seq[int] {
					i := 0
					return For(
						func() bool { return true },
						func() { i++ },
						Combine(Delay[int](func() Seq[int] {
							if i%2 == 0 {
								return Continue[int]()
							}
							if i >= 10 {
								return Break[int]()
							}
							return Normal[int]()
						}), Delay[int](func() Seq[int] {
							return Bind[int](i, func() Seq[int] {
								return Normal[int]()
							})
						})),
					)
				}))
			},
			expect: []int{1, 3, 5, 7, 9},
		},
	}
	for _, it := range all {
		t.Run(it.name, func(t *testing.T) {
			iter := it.factory()
			got := iter2slice(iter)
			assertEqual(t, got, it.expect)
		})
	}
}

func TestSend(t *testing.T) {
	var sent []int
	// for i := 0; ;i++ {
	// 		recv := yield i
	//		println(recv)
	// }
	seq := func() Iterator[int] {
		return Start(Delay(func() Seq[int] {
			i := 0
			return For(
				func() bool { return true },
				func() { i++ },
				Delay(func() Seq[int] {
					var yieldReturn int
					return Combine(Delay(func() Seq[int] {
						return BindRecv(i, func(recv int) Seq[int] {
							yieldReturn = recv
							return Normal[int]()
						})
					}), Delay(func() Seq[int] {
						println(yieldReturn)
						sent = append(sent, yieldReturn)
						return Normal[int]()
					}))
				}),
			)
		}))
	}

	iter := seq()
	g := iter.(*generator[int])
	// g.MoveNext()

	var yieldXS []int

	yieldV, ok := g.Send(42)
	println(yieldV)
	assertEqual(t, ok, true)
	yieldXS = append(yieldXS, yieldV)

	yieldV, ok = g.Send(100)
	println(yieldV)
	assertEqual(t, ok, true)
	yieldXS = append(yieldXS, yieldV)

	assertEqual(t, sent, []int{42, 100})
	assertEqual(t, yieldXS, []int{1, 2})
}

func TestResult(t *testing.T) {
	// yield 1
	// return 42
	seq := func() Iterator[int] {
		return Start(Delay(func() Seq[int] {
			return Bind(1, func() Seq[int] {
				return ReturnValue[int](42)
			})
		}))
	}

	iter := seq()

	got := iter2slice(iter)
	assertEqual(t, got, []int{1})

	g := iter.(*generator[int])
	assertEqual(t, g.Result(), 42)
}

func iter2slice[V any](it Iterator[V]) (xs []V) {
	for it.MoveNext() {
		xs = append(xs, it.Current())
	}
	return xs
}

func assertEqual(t *testing.T, got, expect any) {
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect %+v got %+v", expect, got)
	}
}
