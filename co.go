package co

type Iter[V any] <-chan V

func (*Iter[V]) MoveNext() (_ bool) { return }
func (*Iter[V]) Current() (_ V)     { return }

func Yield[V any](V)           {}
func YieldFrom[V any](Iter[V]) {}
