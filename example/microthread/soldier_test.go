package microthread

import (
	"testing"
	"time"
)

func TestSoldier(t *testing.T) {
	sched := NewSched()
	animateReload := NewSignal()

	soldier := NewSoldier(animateReload)
	soldier.Init(sched)

	for i := 0; ; i++ {
		sched.Run()
		time.Sleep(time.Millisecond * 100)
		if i%50 == 0 {
			animateReload.Set()
		}
	}
}
