package src

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/goghcrow/go-co"
)

func SampleGetNumList() (_ Iter[int]) {
	Yield(1)
	Yield(2)
	Yield(3)
	return
}

func SampleYieldFrom() (_ Iter[int]) {
	Yield(0)
	YieldFrom(SampleGetNumList())
	Yield(4)
	return
}

func SampleForYield() (_ Iter[int]) {
	for i := 0; i < 5; i++ {
		Yield(i)
	}
	return
}

func SampleGetEvenNumbers(start, end int) (_ Iter[int]) {
	for i := start; i < end; i++ {
		if i%2 == 0 {
			Yield(i)
		}
	}
	return
}

func PowersOfTwo(exponent int) (_ Iter[int]) {
	for r, i := 1, 0; i < exponent; i++ {
		Yield(r)
		r *= 2
	}
	return
}

func Fibonacci() Iter[int] {
	a, b := 1, 1
	for {
		Yield(b)
		a, b = b, a+b
	}
}

func Range(start, end, step int) (_ Iter[int]) {
	for i := start; i <= end; i += step {
		Yield(i)
	}
	return
}

func Grep(s string, lines []string) (_ Iter[string]) {
	for _, line := range lines {
		if strings.Contains(line, s) {
			Yield(line)
		}
	}
	return
}

func Run() {
	// Push Mode
	for n := range SampleGetNumList() {
		println(n)
	}

	for n := range SampleYieldFrom() {
		println(n)
	}

	for n := range Fibonacci() {
		if n > 1000 {
			println(n)
			break
		}
	}

	for i := range Range(0, 100, 2) {
		println(i)
	}

	// or Pull Mode

	iter := Range(0, 100, 2)
	for iter.MoveNext() {
		println(iter.Current())
	}
}

type Line struct {
	Bytes  []byte
	Prefix bool
	Err    error
}

func ReadFile(name string) (_ Iter[Line]) {
	file, err := os.Open(name)
	// defer file.Close()
	if err != nil {
		Yield(Line{Err: err})
		return
	}

	r := bufio.NewReader(file)
	for {
		line, prefix, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			Yield(Line{Err: err})
		} else {
			Yield(Line{Bytes: line, Prefix: prefix, Err: err})
		}
	}
	return
}

func TestSample(t *testing.T) {
	all := []struct {
		name    string
		factory func() Iter[int]
		expect  []int
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
			name:    "SampleForYield",
			factory: SampleForYield,
			expect:  []int{0, 1, 2, 3, 4},
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
			expect: []int{10, 12, 14, 16, 18, 20},
		},
	}
	for _, it := range all {
		t.Run(it.name, func(t *testing.T) {
			got := iter2slice(it.factory())
			assertEqual(t, got, it.expect)
		})
	}
}
