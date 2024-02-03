package microthread

// Signal
// a simple synchronization primitive let the micro-thread obtain and yield a
// signal object, which means it will not be scheduled until the signal has been
// set. Instead of using blocking APIs, you can use async APIs, create a signal
// object, wait on that, and have the callback set the signal. Or, some game
// object's controlling micro-thread might want to sleep until another game
// object reaches a certain state; the target object can keep a signal
// accessible via a property that micro-threads can read and wait on.
//
// A thread can wait for more than one signal, and more than one thread can wait
// for a signal.
//
// The signal’s job is to keep a list of all the tasks that are waiting for it.
// When tasks start waiting, they move out of the scheduler’s lists and are
// tracked by all the signals instead. Each signal increments/decrements the
// TaskItem's waitForSigCnt field to keep track of how many signals the task is
// waiting for. When the count reaches zero, the task can be moved back to the
// scheduler.
type Signal struct {
	// id    int
	tasks []*TaskItem
	isSet bool
}

func NewSignal() *Signal {
	return &Signal{isSet: true}
}

// Update
// when the task returns a signal or slice of signals, it's
// moved from the scheduler's lists to the tasks' lists.
func (s *Signal) Update(sched *Sched, it *TaskIter) {
	t := it.RemoveCurrent()
	t.waitForSigCnt = 0
	s.add(t)
}

func (s *Signal) add(task *TaskItem) {
	s.isSet = false
	s.tasks = append(s.tasks, task)
	task.waitForSigCnt++
}

func (s *Signal) Set() {
	if s.isSet {
		return
	}
	s.isSet = true

	for _, task := range s.tasks {
		task.waitForSigCnt--
		if task.waitForSigCnt == 0 {
			task.sched.addToActive(task)
		}
	}
	s.tasks = nil
}

type Signals []*Signal

func NewSignals(xs ...*Signal) Signals {
	return xs
}

func (s *Signals) Update(sched *Sched, it *TaskIter) {
	for _, sig := range *s {
		sig.Update(sched, it)
	}
}
