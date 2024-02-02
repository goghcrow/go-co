//go:build co

package sched

import (
	"fmt"
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

func Sleep[Ctx any](d time.Duration) AsyncFn[Ctx] {
	return func(_ Ctx, k Continuation) {
		timeAfter(d, k)
	}
}

type SampleCtx struct {
	Result int
}

func SampleAsyncTask(v int) AsyncFn[*SampleCtx] {
	return func(ctx *SampleCtx, k Continuation) {
		timeAfter(time.Second*1, func() {
			ctx.Result = v
			k()
		})
	}
}

func TestCo(t *testing.T) {
	defer DeferMain()

	echo := func(f string, v ...any) {
		now := time.Now().Format("15:04:05")
		fmt.Printf("[%s] "+f+"\n", append([]any{now}, v...)...)
	}

	AsyncRun(func() (_ Iter[OnCompleted]) {
		echo("start")

		ctx := &SampleCtx{}
		Yield(Await(ctx, Sleep[*SampleCtx](time.Second*1)))
		echo("after sleep")

		Yield(Await(ctx, SampleAsyncTask(42)))
		echo("done, result is %d", ctx.Result)
		return
	})
}
