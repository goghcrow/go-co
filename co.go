package co

// Iter is a 𝗦𝘆𝗻𝘁𝗮𝗰𝘁𝗶𝗰 𝗦𝘂𝗴𝗮𝗿
// please coding depending on the type parameter [V]
// instead of the underlying type <-chan
type Iter[V any] <-chan V

func (Iter[V]) MoveNext() (_ bool) { return }
func (Iter[V]) Current() (_ V)     { return }

func Yield[V any](V)           {}
func YieldFrom[V any](Iter[V]) {}
