//go:build co

package sched

import (
	"testing"
	"time"

	. "github.com/goghcrow/go-co"
)

// fake callback
func timeAfter(d time.Duration, cb func()) {
	go func() {
		time.Sleep(d)
		cb()
	}()
}

func Sleep(d time.Duration) Async {
	return AsyncFun(func(cont func(v any, err error)) {
		timeAfter(d, func() {
			cont(nil, nil)
		})
	})
}

func SampleAsyncTask(v any) Async {
	return AsyncFun(func(cont func(v any, err error)) {
		timeAfter(time.Second*1, func() {
			cont(v, nil)
		})
	})
}

func TestCo(t *testing.T) {
	now := func() string { return time.Now().Format("2006-01-02 15:04:05") }

	Co(func(s *Sched) (_ Iter[Async]) {
		t.Log("start")

		t.Log(now() + " before sleep")
		Yield(Sleep(time.Second * 1))

		t.Log(now() + " before async task")
		Yield(SampleAsyncTask(42))

		t.Log(now() + " after async task and get result")
		result, _ := s.GetReceive()
		t.Log(result)

		t.Log("end")
		return
	})

	wg.Wait()
}
