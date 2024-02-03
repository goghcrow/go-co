package microthread

type TaskList struct {
	hd *TaskItem
	tl *TaskItem
}

type TaskIter struct {
	list       *TaskList
	curr, prev *TaskItem
}

func (l *TaskList) Append(item *TaskItem) {
	assert(item.next == nil)
	if l.hd == nil {
		l.hd = item
		l.tl = item
	} else {
		l.tl.next = item
		l.tl = item
	}
}

func (l *TaskList) Remove(item, prev *TaskItem) {
	if prev == nil {
		assert(l.hd == item)
		l.hd = item.next
	} else {
		assert(prev.next == item)
		prev.next = item.next
	}
	if item.next == nil {
		assert(l.tl == item)
		l.tl = prev
	}
	item.next = nil
}

func (l *TaskList) GetIterator() *TaskIter {
	return &TaskIter{list: l}
}

func (i *TaskIter) Current() *TaskItem {
	return i.curr
}

func (i *TaskIter) MoveNext() bool {
	var next *TaskItem
	if i.curr == nil {
		if i.prev == nil {
			next = i.list.hd
		} else {
			// current is removed, don't need to update prev
			next = i.prev.next
		}
	} else {
		next = i.curr.next
	}

	if next == nil {
		return false
	}

	if i.curr != nil {
		i.prev = i.curr
	}
	i.curr = next
	return true
}

func (i *TaskIter) MoveCurrentToList(l *TaskList) {
	l.Append(i.RemoveCurrent())
}

func (i *TaskIter) RemoveCurrent() (curr *TaskItem) {
	assert(i.curr != nil)
	curr = i.curr
	i.list.Remove(i.curr, i.prev)
	i.curr = nil
	return
}
