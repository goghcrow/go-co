package seq

import "reflect"

// https://go.dev/ref/spec#For_statements
// A "for" statement with a "range" clause iterates through all entries of
// an array, slice, string or map, or values received on a channel.

type Pair[K, V any] struct {
	Key K
	Val V
}

type MapIter[K comparable, V any] struct {
	iter *reflect.MapIter
}

func NewMapIter[K comparable, V any](m map[K]V) Iterator[Pair[K, V]] {
	return &MapIter[K, V]{
		iter: reflect.ValueOf(m).MapRange(),
	}
}

func (m *MapIter[K, V]) MoveNext() bool {
	return m.iter.Next()
}

func (m *MapIter[K, V]) Current() Pair[K, V] {
	return Pair[K, V]{
		Key: m.iter.Key().Interface().(K),
		Val: m.iter.Value().Interface().(V),
	}
}

type SliceIter[V any] struct {
	slice []V
	idx   int
}

func NewSliceIter[V any](slice []V) Iterator[Pair[int, V]] {
	return &SliceIter[V]{slice: slice, idx: -1}
}

func (s *SliceIter[V]) MoveNext() bool {
	s.idx++
	return s.idx < len(s.slice)
}

func (s *SliceIter[V]) Current() Pair[int, V] {
	return Pair[int, V]{Key: s.idx, Val: s.slice[s.idx]}
}

type ArrayIter[V any] struct {
	array reflect.Value
	idx   int
}

func NewArrayIter[V any](array any) Iterator[Pair[int, V]] {
	return &ArrayIter[V]{array: reflect.ValueOf(array), idx: -1}
}

func (a *ArrayIter[V]) MoveNext() bool {
	a.idx++
	return a.idx < a.array.Len()
}

func (a *ArrayIter[V]) Current() Pair[int, V] {
	return Pair[int, V]{Key: a.idx, Val: a.array.Index(a.idx).Interface().(V)}
}

type StringIter struct {
	str string
	idx int
}

func NewStringIter(str string) Iterator[Pair[int, byte]] {
	return &StringIter{str: str, idx: -1}
}

func (s *StringIter) MoveNext() bool {
	s.idx++
	return s.idx < len(s.str)
}

func (s *StringIter) Current() Pair[int, byte] {
	return Pair[int, byte]{Key: s.idx, Val: s.str[s.idx]}
}

type ChanIter[V any] struct {
	ch chan V
	v  V
}

func NewChanIter[V any](ch chan V) Iterator[V] {
	return &ChanIter[V]{ch: ch}
}

func (c *ChanIter[V]) MoveNext() (ok bool) {
	c.v, ok = <-c.ch
	return
}

func (c *ChanIter[V]) Current() V {
	return c.v
}
