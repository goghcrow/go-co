//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen
//go:generate cogen

package example

import (
	"bufio"
	"io"
	"os"
	"strings"

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

type Pair[K, V any] struct {
	Key K
	Val V
}

func SampleLoop() (_ Iter[any]) {
	for i := 0; i < 5; i++ {
		Yield(i)
	}

	xs := []string{"a", "b", "c"}
	for i, n := range xs {
		Yield(Pair[int, string]{i, n})
	}

	for i, c := range "Hello World!" {
		Yield(Pair[int, rune]{i, c})
	}
	return
}

func SampleLoopMap() (_ Iter[Pair[string, int]]) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	for k, v := range m {
		Yield(Pair[string, int]{k, v})
	}
	return nil
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
