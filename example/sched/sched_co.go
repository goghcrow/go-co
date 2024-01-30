//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen
//go:generate cogen

package sched

import (
	"sync"

	. "github.com/goghcrow/go-co"
)

var wg = &sync.WaitGroup{}

type Sched struct {
	val any
	err error
}

func (s *Sched) run(co func(s *Sched) Iter[Async]) {
	it := co(s)

	var run func()
	run = func() {
		for it.MoveNext() {
			switch v := it.Current().(type) {
			case Async:
				v.Begin(func(v any, err error) {
					s.send(v, err)
					run()
				})
				return
			default:
				panic("unreached")
			}
		}
		wg.Done()
	}

	wg.Add(1)
	run()
}

func (s *Sched) send(v any, err error) {
	s.val = v
	s.err = err
}

func (s *Sched) GetReceive() (any, error) {
	return s.val, s.err
}

func Co(co func(s *Sched) Iter[Async]) {
	sched := &Sched{}
	sched.run(co)
}

type Async interface {
	Begin(cont func(v any, err error))
}

type AsyncFun func(cont func(v any, err error))

func (f AsyncFun) Begin(cont func(v any, err error)) {
	f(cont)
}
