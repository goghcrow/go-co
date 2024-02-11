//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen@main
//go:generate cogen

package sched

import (
	"sync"

	. "github.com/goghcrow/go-co"
)

type (
	Act              func()
	Act1[T1 any]     func(T1)
	Act2[T1, T2 any] func(T1, T2)

	Continuation     Act
	OnCompleted      Act1[Continuation]
	AsyncFn[Ctx any] Act2[Ctx, Continuation]
)

var wg = &sync.WaitGroup{}

func DeferMain() {
	wg.Wait()
}

func Await[T any](ctx T, f AsyncFn[T]) OnCompleted {
	return func(k Continuation) {
		f(ctx, k)
	}
}

func AsyncRun(f func() Iter[OnCompleted]) {
	it := f()

	var recur func()
	recur = func() {
		if it.MoveNext() {
			it.Current()(recur)
			return
		}
		wg.Done()
	}

	wg.Add(1)
	recur()
}
