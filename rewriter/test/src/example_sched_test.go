package src

import (
	"sync"
	"testing"
	"time"

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

// ------------------------------------------------------------

func Sleep(d time.Duration) Async {
	return AsyncFun(func(cont func(v any, err error)) {
		timeAfter(d, func() {
			cont(nil, nil)
		})
	})
}

func SampleAsyncTask(v any) Async {
	return AsyncFun(func(cont func(v any, err error)) {
		timeAfter(time.Second, func() {
			cont(v, nil)
		})
	})
}

func TestCo(t *testing.T) {
	Co(func(s *Sched) (_ Iter[Async]) {
		t.Log("start")

		t.Log(now() + " before sleep")
		Yield(Sleep(time.Second * 1))

		t.Log(now() + " before async task")
		Yield(SampleAsyncTask(42))

		t.Log(now() + " after async task")

		t.Log(now() + " get async task result")
		result, _ := s.GetReceive()
		t.Log(result)

		t.Log("end")
		return
	})

	wg.Wait()
}

// ------------------------------------------------------------

// fake callback
func timeAfter(d time.Duration, cb func()) {
	go func() {
		time.Sleep(d)
		cb()
	}()
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
