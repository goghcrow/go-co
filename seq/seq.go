package seq

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Monadic 🆈🅸🅴🅻🅳 Implementation
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type contType int

const (
	kNormal contType = iota
	kBreak
	kContinue
	kReturn
)

type (
	step[V any] struct {
		value V       // current value
		next  next[V] // compute the next step
	}
	next[V any]     func(recv V) *step[V]   // the next step computation
	co[V any]       struct{ step *step[V] } // coroutine, which stores the current value and the next step
	lazy[V any]     func() Seq[V]           // thunk, boxing code after yield for later execution
	lazyRecv[V any] func(recv V) Seq[V]     // with receive value
	cont[V any]     func(contType, V)       // continuation
)

type (
	Seq[V any]      func(*co[V] /*state*/, cont[V]) // Async Sequence / Async Enumerator
	Iterator[V any] interface {
		MoveNext() bool
		Current() V
	}
	Generator[V any] interface {
		Iterator[V]
		Result() V
		Send(V) (yield V, ok bool)
	}
)

// type Iterable[V any] interface { GetIterator() Iterator[V] }
// type IterableFn[V any] func() Iterator[V]
// func (f IterableFn[V]) GetIterator() Iterator[V] { return f() }

func zero[V any]() (z V) { return }

func mkNextRecv[V any](f lazyRecv[V], c *co[V], k cont[V]) next[V] {
	return func(recv V) *step[V] {
		// c.step = nil
		f(recv)(c, k) // compute next, set the step if bind called,
		s := c.step   // otherwise nil
		c.step = nil
		return s
	}
}

func mkNext[V any](f lazy[V], c *co[V], k cont[V]) next[V] {
	return mkNextRecv[V](func(_ V) Seq[V] { return f() }, c, k)
}

// Start / Run a coroutine (Delimited Continuation) in boundary
func Start[V any](seq Seq[V]) Iterator[V] {
	var it *generator[V]
	it = newGenerator[V](mkNext(
		func() Seq[V] { return seq },
		&co[V]{},
		func(t contType, v V) { it.result = v },
	))
	return it
}

// Bind collect pending stack frame,
// When Bind() called, saving k to co.step.next and return immediately
// When generator.MoveNext() called, co.step.next will be invoked
// supporting yield statement without return value ( <- iter.Send())
func Bind[V any](v V, f lazy[V]) Seq[V] {
	return func(c *co[V], k cont[V]) {
		c.step = &step[V]{
			value: v,
			next:  mkNext(f, c, k),
		}
	}
}

// BindRecv with return value
// supporting yield expression with return value
func BindRecv[V any](v V, f lazyRecv[V]) Seq[V] {
	return func(c *co[V], k cont[V]) {
		c.step = &step[V]{
			value: v,
			next:  mkNextRecv(f, c, k),
		}
	}
}

func For[V any](
	cond func() bool,
	post func(),
	body Seq[V],
) Seq[V] {
	return func(c *co[V], k cont[V]) {
		var loop func(skipPost bool)
		loop = func(skipPost bool) {
			if post != nil && !skipPost {
				post()
			}
			if cond == nil || cond() {
				body(c, func(t contType, v V) {
					switch t {
					case kNormal, kContinue:
						loop(false)
					case kBreak:
						k(kNormal, zero[V]())
					case kReturn:
						k(kReturn, v)
					default:
						panic("unreachable")
					}
				})
			} else {
				k(kNormal, zero[V]())
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

func Delay[V any](f lazy[V]) Seq[V] {
	return func(c *co[V], k cont[V]) {
		f()(c, k)
	}
}

func Combine[V any](s1, s2 Seq[V]) Seq[V] {
	return func(c *co[V], k cont[V]) {
		s1(c, func(t contType, v V) {
			// skip s2 when break/continue/return
			// notice: break/continue
			// whether the outer is loop or not
			if t == kNormal {
				s2(c, k)
			} else {
				k(t, v)
			}
		})
	}
}

func seqOfK[V any](kt contType) Seq[V] {
	return func(c *co[V], k cont[V]) {
		k(kt, zero[V]())
	}
}

func Normal[V any]() Seq[V]   { return seqOfK[V](kNormal) }
func Break[V any]() Seq[V]    { return seqOfK[V](kBreak) }
func Continue[V any]() Seq[V] { return seqOfK[V](kContinue) }
func Return[V any]() Seq[V]   { return seqOfK[V](kReturn) }
func ReturnValue[V any](v V) Seq[V] { // supporting generator with return value
	return func(c *co[V], k cont[V]) {
		k(kReturn, v)
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 🅶🅴🅽🅴🆁🅰🆃🅾🆁 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// asyncIter
type generator[V any] struct {
	started bool
	next    next[V]
	current/*, ok*/ V
	result/*, ok*/ V
}

func newGenerator[V any](next next[V]) *generator[V] {
	return &generator[V]{next: next}
}

func (d *generator[V]) Result() V {
	return d.result
}

func (d *generator[V]) Current() V {
	// assert(d.started)
	return d.current
}

func (d *generator[V]) MoveNext() bool {
	d.started = true
	return d.moveNext(zero[V]())
}

func (d *generator[V]) Send(v V) (V, bool) {
	if !d.started {
		// if the generator is not at a yield expression when this method is called,
		// it will first be let to advance to the first yield expression before sending the value.
		// so, the first current value would be skipped also
		if !d.MoveNext() {
			return zero[V](), false
		}
	}
	if d.moveNext(v) {
		return d.current, true
	} else {
		return zero[V](), false
	}
}

func (d *generator[V]) moveNext(sent V) bool {
	if d.next == nil {
		return false
	}
	s := d.next(sent) // compute next step
	if s == nil {
		d.next = nil
		d.current = zero[V]()
		return false
	} else {
		d.next = s.next
		d.current = s.value
		return true
	}
}
