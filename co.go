package co

// syntactic sugar interface declaration

// please code depending on the type parameter instead of
// the concrete and underlying type, e.g. <-chan

type Iter[V any] <-chan V

func (*Iter[V]) MoveNext() (_ bool) { return }
func (*Iter[V]) Current() (_ V)     { return }

// func (*Iter[V]) Send(V) (yield V, ok bool) { return }
// func (*Iter[V]) Result() (_ V)             { return }

func Yield[V any](V)           {}
func YieldFrom[V any](Iter[V]) {}
