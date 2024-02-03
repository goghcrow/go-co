package microthread

import (
	"time"
)

type Sched struct {
	active, sleeping *TaskList
	now              time.Time
}

func NewSched() *Sched {
	return &Sched{
		active:   &TaskList{},
		sleeping: &TaskList{},
	}
}

func (s *Sched) addToActive(task *TaskItem) {
	s.active.Append(task)
}

func (s *Sched) newTaskItem(task Task) *TaskItem {
	return &TaskItem{sched: s, Task: task}
}

func (s *Sched) AddTask(task Task) {
	s.active.Append(s.newTaskItem(task))
}

func (s *Sched) Run() {
	s.now = time.Now()

	it := s.sleeping.GetIterator()
	for it.MoveNext() {
		if s.now.After(it.Current().wakeUp) {
			it.MoveCurrentToList(s.active)
		}
	}

	it = s.active.GetIterator()
	for it.MoveNext() {
		c := it.Current()
		if !c.Task.MoveNext() {
			it.RemoveCurrent()
			continue
		}
		state := c.Task.Current()
		if state != nil {
			state.Update(s, it)
		}
	}
}
