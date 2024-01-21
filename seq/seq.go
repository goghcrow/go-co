package seq

// Monadic Yield Implementation

type ContType int

const (
	KNormal ContType = iota
	KBreak
	KContinue
	KReturn
)

type (
	Step[V any] struct {
		value V
		next  func() *Step[V]
	}
	Co[V any]   struct{ step *Step[V] } // Coroutine
	Cont[V any] func(ContType, V)       // Continuation
	Seq[V any]  func(*Co[V], Cont[V])   // Async Sequence / Async Enumerator
)

type (
	Iterator[V any] interface {
		MoveNext() bool
		Current() V
	}
	// Iterable[V any] interface{ GetIter() Iterator[V] }
)

// Start / Run / Bind (Yield) Boundary
// **Delimited** Continuation
func Start[V any](step Seq[V]) Iterator[V] {
	c := &Co[V]{}
	return NewDelegateIter[V](func() *Step[V] {
		c.step = nil
		step(c, func(t ContType, v V) {})
		s := c.step
		c.step = nil
		return s
	})
}

// Bind collect pending stack frame when invoking
// When Bind() called, saving k to Co.step.next and return immediately
// When DelegateIter.MoveNext() called, Co.step.next will be invoked
func Bind[V any](v V, f func() Seq[V]) Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		c.step = &Step[V]{
			value: v,
			next: func() *Step[V] {
				c.step = nil
				f()(c, k)
				s := c.step
				c.step = nil
				return s
			},
		}
	}
}

func For[V any](
	cond func() bool,
	post func(),
	body Seq[V],
) Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		var loop func(skipPost bool)
		loop = func(skipPost bool) {
			if post != nil && !skipPost {
				post()
			}
			if cond == nil || cond() {
				body(c, func(t ContType, v V) {
					switch t {
					case KNormal, KContinue:
						loop(false)
					case KBreak:
						k(KNormal, zero[V]())
					case KReturn:
						k(KReturn, v)
					default:
						panic("unreachable")
					}
				})
			} else {
				k(KNormal, zero[V]())
			}
		}
		loop(true)
	}
}

func While[V any](cond func() bool, body Seq[V]) Seq[V] {
	return For(cond, nil, body)
}

func Loop[V any](body Seq[V]) Seq[V] {
	return For(nil, nil, body)
}

func Range[V any](cond func() bool, body Seq[V]) Seq[V] {
	return For(cond, nil, body)
}

func Delay[V any](f func() Seq[V]) Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		f()(c, k)
	}
}

func Combine[V any](s1, s2 Seq[V]) Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		s1(c, func(t ContType, v V) {
			if t == KNormal {
				s2(c, k)
			} else {
				k(t, v)
			}
		})
	}
}

func Return[V any]( /*v V*/ ) Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		v := zero[V]()
		k(KReturn, v)
	}
}

func Normal[V any]() Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		k(KNormal, zero[V]())
	}
}

func Break[V any]() Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		k(KBreak, zero[V]())
	}
}

func Continue[V any]() Seq[V] {
	return func(c *Co[V], k Cont[V]) {
		k(KContinue, zero[V]())
	}
}

func zero[V any]() (z V) { return }

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Iterator ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

type DelegateIter[V any] struct {
	delegate func() *Step[V]
	current  V
}

// NewDelegateIter NewAsyncIter
func NewDelegateIter[V any](delegate func() *Step[V]) *DelegateIter[V] {
	return &DelegateIter[V]{delegate: delegate}
}

func (d *DelegateIter[V]) MoveNext() bool {
	if d.delegate == nil {
		// panic("illegal state")
		return false
	}
	step := d.delegate()
	if step == nil {
		d.delegate = nil
		d.current = zero[V]()
		return false
	} else {
		d.delegate = step.next
		d.current = step.value
		return true
	}
}

func (d *DelegateIter[V]) Current() V {
	return d.current
}
