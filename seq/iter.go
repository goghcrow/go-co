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

type StringIter struct {
	str []rune
	idx int
}

func NewStringIter(str string) Iterator[Pair[int, rune]] {
	return &StringIter{str: []rune(str), idx: -1}
}

func (s *StringIter) MoveNext() bool {
	s.idx++
	return s.idx < len(s.str)
}

func (s *StringIter) Current() Pair[int, rune] {
	return Pair[int, rune]{Key: s.idx, Val: s.str[s.idx]}
}

type IntegerIter struct {
	n int
	i int
}

func NewIntegerIter(n int) Iterator[Pair[int, any]] {
	return &IntegerIter{n: n}
}

func (i *IntegerIter) MoveNext() bool {
	i.i++
	return i.i <= i.n
}

func (i *IntegerIter) Current() Pair[int, any] {
	return Pair[int, any]{Key: i.i}
}

type ChanIter[V any] struct {
	ch <-chan V
	v  V
}

func NewChanIter[V any](ch <-chan V) Iterator[Pair[V, any]] {
	return &ChanIter[V]{ch: ch}
}

func (c *ChanIter[V]) MoveNext() (ok bool) {
	c.v, ok = <-c.ch
	return
}

func (c *ChanIter[V]) Current() Pair[V, any] {
	return Pair[V, any]{Key: c.v}
}
