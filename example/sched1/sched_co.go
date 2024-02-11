//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen@main
//go:generate cogen

package sched

import (
	"sync"

	. "github.com/goghcrow/go-co"
)

// Async interface { Begin(Continuation) }
// type AsyncFun func(Continuation)
// func (f AsyncFun) Begin(k Continuation) { f(k) }

type (
	Continuation func(v any, err error)
	Async        func(Continuation)
	Sched        struct {
		val any
		err error
	}
)

func Co(co func(s *Sched) Iter[Async]) {
	sched := &Sched{}
	sched.run(co)
}

var wg = &sync.WaitGroup{}

func DeferMain() {
	wg.Wait()
}

func (s *Sched) run(co func(s *Sched) Iter[Async]) {
	it := co(s)

	var recur func()
	recur = func() {
		if it.MoveNext() {
			it.Current()(func(v any, err error) {
				s.Send(v, err)
				recur()
			})
			return
		}
		wg.Done()
	}

	wg.Add(1)
	recur()
}

func (s *Sched) Send(v any, err error) {
	s.val = v
	s.err = err
}

func (s *Sched) GetReceive() (any, error) {
	return s.val, s.err
}
