package microthread

import (
	"time"

	"github.com/goghcrow/go-co/seq"
)

type Task = seq.Iterator[State]

type TaskItem struct {
	sched *Sched

	wakeUp        time.Time // sleep until
	waitForSigCnt int       // how many signals the task is waiting for

	Task Task
	next *TaskItem
}
