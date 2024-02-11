//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen@main
//go:generate cogen

package linq

import (
	. "github.com/goghcrow/go-co"
)

func Of[A any](xs ...A) (_ Iter[A]) {
	for _, x := range xs {
		Yield(x)
	}
	return
}

func Range(start, end int, stepOpt ...int) (_ Iter[int]) {
	step := 1
	if len(stepOpt) > 0 {
		step = stepOpt[0]
	}
	for i := start; i < end; i += step {
		Yield(i)
	}
	return
}

// ========================================================

type (
	Selector[A, R any] func(A) R
	Predicate[A any]   func(A) bool
)

// Unit / Return / Pure
func Unit[A any](a A) (_ Iter[A]) {
	Yield(a)
	return
}

// SelectMany / Bind / FlatMap
func SelectMany[A, R any](it Iter[A], f func(A) Iter[R]) (_ Iter[R]) {
	for a := range it {
		YieldFrom(f(a))
	}
	return
}

func Select[A, R any](it Iter[A], f Selector[A, R]) (_ Iter[R]) {
	for a := range it {
		Yield(f(a))
	}
	return
}

func Where[A any](it Iter[A], p Predicate[A]) (_ Iter[A]) {
	for a := range it {
		if p(a) {
			Yield(a)
		}
	}
	return
}

func First[A any](it Iter[A]) (a A, has bool) {
	if it.MoveNext() {
		return it.Current(), true
	}
	return
}

func FirstWhile[A any](it Iter[A], p Predicate[A]) (fst A, has bool) {
	for a := range it {
		if p(a) {
			return a, true
		}
	}
	return
}

func Last[A any](it Iter[A]) (last A, has bool) {
	for a := range it {
		last = a
		has = true
	}
	return
}

func LastWhile[A any](it Iter[A], p Predicate[A]) (last A, has bool) {
	for a := range it {
		if p(a) {
			last = a
			has = true
		}
	}
	return
}

func Take[A any](it Iter[A], cnt int) (_ Iter[A]) {
	for a := range it {
		if cnt <= 0 {
			break
		}
		cnt--
		Yield(a)
	}
	return
}

func TakeWhile[A any](it Iter[A], p Predicate[A]) (_ Iter[A]) {
	return Where(it, p)
}

func Skip[A any](it Iter[A], cnt int) (_ Iter[A]) {
	for a := range it {
		if cnt > 0 {
			cnt--
			continue
		}
		Yield(a)
	}
	return
}

func SkipWhile[A any](it Iter[A], p Predicate[A]) (_ Iter[A]) {
	for a := range it {
		if p(a) {
			continue
		}
		Yield(a)
	}
	return
}

func Aggregate[A, B, R any](
	it Iter[A],
	init B,
	f func(acc B, cur A) B,
	selector Selector[B, R],
) (r R) {
	acc := init
	for a := range it {
		acc = f(acc, a)
	}
	return selector(acc)
}

func Fold[A, R any](
	it Iter[A],
	init R,
	f func(acc R, cur A) R,
) (acc R) {
	acc = init
	for a := range it {
		acc = f(acc, a)
	}
	return
}

func Reduce[A any](
	it Iter[A],
	f func(acc A, cur A) A,
) (acc A, ok bool) {
	if !it.MoveNext() {
		return
	}
	acc = it.Current()
	for a := range it {
		acc = f(acc, a)
		ok = true
	}
	return
}

func All[A any](it Iter[A], p Predicate[A]) bool {
	for a := range it {
		if !p(a) {
			return false
		}
	}
	return true
}

func AnyElem[A any](it Iter[A]) bool {
	return it.MoveNext()
}

func Any[A any](it Iter[A], p Predicate[A]) bool {
	for a := range it {
		if p(a) {
			return true
		}
	}
	return false
}

func Append[A any](it Iter[A], a A) (_ Iter[A]) {
	YieldFrom(it)
	Yield(a)
	return
}
