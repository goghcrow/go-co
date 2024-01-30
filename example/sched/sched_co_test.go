//go:build co

package sched

import (
	"testing"
	"time"

	. "github.com/goghcrow/go-co"
)

func TestCo(t *testing.T) {
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
