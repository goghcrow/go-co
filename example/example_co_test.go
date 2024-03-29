//go:build co

package example

import (
	"reflect"
	"testing"

	. "github.com/goghcrow/go-co"
)

func TestSampleMap(t *testing.T) {
	it := SampleLoopMap()
	m := map[string]int{}
	for it.MoveNext() {
		m[it.Current().Key] = it.Current().Val
	}
	assertEqual(t, len(m), 3)
	assertEqual(t, m["a"], 1)
	assertEqual(t, m["b"], 2)
	assertEqual(t, m["c"], 3)
}

func TestSample(t *testing.T) {
	all := []struct {
		name    string
		factory any
		expect  any
	}{
		{
			name:    "SampleGetNumList",
			factory: SampleGetNumList,
			expect:  []int{1, 2, 3},
		},
		{
			name:    "SampleYieldFrom",
			factory: SampleYieldFrom,
			expect:  []int{0, 1, 2, 3, 4},
		},
		{
			name:    "SampleLoop",
			factory: SampleLoop,
			expect: []any{
				0, 1, 2, 3, 4,
				Pair[int, string]{0, "a"},
				Pair[int, string]{1, "b"},
				Pair[int, string]{2, "c"},
				Pair[int, rune]{0, 'H'},
				Pair[int, rune]{1, 'e'},
				Pair[int, rune]{2, 'l'},
				Pair[int, rune]{3, 'l'},
				Pair[int, rune]{4, 'o'},
				Pair[int, rune]{5, ' '},
				Pair[int, rune]{6, 'W'},
				Pair[int, rune]{7, 'o'},
				Pair[int, rune]{8, 'r'},
				Pair[int, rune]{9, 'l'},
				Pair[int, rune]{10, 'd'},
				Pair[int, rune]{11, '!'},
			},
		},
		{
			name: "SampleGetEvenNumbers",
			factory: func() Iter[int] {
				return SampleGetEvenNumbers(0, 10)
			},
			expect: []int{0, 2, 4, 6, 8},
		},
		{
			name: "PowersOfTwo",
			factory: func() Iter[int] {
				return PowersOfTwo(5)
			},
			expect: []int{1, 2, 4, 8, 16},
		},
		{
			name: "Fibonacci",
			factory: func() (_ Iter[int]) {
				for n := range Fibonacci() {
					if n > 1000 {
						Yield(n)
						break
					}
				}
				return
			},
			expect: []int{1597},
		},
		{
			name: "Range",
			factory: func() Iter[int] {
				return Range(10, 20, 2)
			},
			expect: []int{10, 12, 14, 16, 18},
		},
	}
	for _, it := range all {
		t.Run(it.name, func(t *testing.T) {
			switch f := it.factory.(type) {
			case func() Iter[int]:
				got := iter2slice(f())
				assertEqual(t, got, it.expect)
			case func() Iter[any]:
				got := iter2slice(f())
				assertEqual(t, got, it.expect)
			}
		})
	}
}

func iter2slice[V any](g Iter[V]) (xs []V) {
	for i := range g {
		xs = append(xs, i)
	}
	return xs
}

func assertEqual(t *testing.T, got, expect any) {
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect\n%+v\n\ngot\n%+v", expect, got)
	}
}
