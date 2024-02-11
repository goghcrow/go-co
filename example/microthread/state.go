package microthread

import "time"

// State ThreadState
type State interface {
	Update(*Sched, *TaskIter)
}

type StateFn func(*Sched, *TaskIter)

func (f StateFn) Update(s *Sched, it *TaskIter) { f(s, it) }

// WaitFor
// when the task returns WaitFor, move to the sleeping list
func WaitFor(d time.Duration) State {
	assert(d >= 0)
	return StateFn(func(s *Sched, it *TaskIter) {
		it.Current().wakeUp = s.now.Add(d)
		it.MoveCurrentToList(s.sleeping)
	})
}
