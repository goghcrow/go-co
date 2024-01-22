## ROADMAP

Generator can yield values and also can return a single value.


```golang
package co

type Iter[V, R any] <-chan V

func Yield[V any](V)           {}
func YieldFrom[V, R any](Iter[V, R]) {}
func Return[V, R any](R) Iter[V, R] { return nil }
```

```golang
package seq

type Cont[V any] func(ContType, V/*ReturnValue*/)

type Iterator[V, R any] interface {
    MoveNext() bool
    Current() V
    GetResult() R/*ReturnValue*/
	// Send()
}

func Return[V any](v V/*ReturnValue*/) Seq[V] {
    return func(c *Co[V], k Cont[V]) {
        k(KReturn, v)
    }
}

func Start[V any](step Seq[V]) Iterator[V] {
    c := &Co[V]{}
    return NewDelegateIter[V](func() *Step[V] {
        c.step = nil
        step(c, func(t ContType, v V) { /*v is ReturnValue*/ })
        s := c.step
        c.step = nil
        return s
    })
}

type Step[V, R any] struct {
    value V
	result R
    next  func() *Step[V, R]
} 

type DelegateIter[V, R any] struct {
	delegate func() *Step[V, R]
	current  V
	result R
}

```

```golang
package test

func Range(min, max int) Iter[int, int] {
	sum := 0
	for i := min; i < max; i++ {
		Yield(i)
	}
	return Return(sum)
}
```