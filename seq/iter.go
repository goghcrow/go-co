package seq

import "reflect"

// helper for rewrite for range statement

// https://go.dev/ref/spec#For_statements
// A "for" statement with a "range" clause iterates through all entries of
// an array, slice, string or map, or values received on a channel.

type pair[K, V any] struct {
	Key K
	Val V
}

func NewIntegerIter(n int) Iterator[pair[int, any]] {
	return &integerIter{n: n}
}

func NewStringIter(str string) Iterator[pair[int, rune]] {
	return &stringIter{str: []rune(str), idx: -1}
}

func NewSliceIter[V any](slice []V) Iterator[pair[int, V]] {
	return &sliceIter[V]{slice: slice, idx: -1}
}

func NewMapIter[K comparable, V any](m map[K]V) Iterator[pair[K, V]] {
	return &mapIter[K, V]{
		iter: reflect.ValueOf(m).MapRange(),
	}
}

func NewChanIter[V any](ch <-chan V) Iterator[pair[V, any]] {
	return &chanIter[V]{ch: ch}
}

type integerIter struct {
	n int
	i int
}

func (i *integerIter) MoveNext() bool {
	i.i++
	return i.i <= i.n
}

func (i *integerIter) Current() pair[int, any] {
	return pair[int, any]{Key: i.i}
}

type stringIter struct {
	str []rune
	idx int
}

func (s *stringIter) MoveNext() bool {
	s.idx++
	return s.idx < len(s.str)
}

func (s *stringIter) Current() pair[int, rune] {
	return pair[int, rune]{Key: s.idx, Val: s.str[s.idx]}
}

type sliceIter[V any] struct {
	slice []V
	idx   int
}

func (s *sliceIter[V]) MoveNext() bool {
	s.idx++
	return s.idx < len(s.slice)
}

func (s *sliceIter[V]) Current() pair[int, V] {
	return pair[int, V]{Key: s.idx, Val: s.slice[s.idx]}
}

type mapIter[K comparable, V any] struct {
	iter *reflect.MapIter
}

func (m *mapIter[K, V]) MoveNext() bool {
	return m.iter.Next()
}

func (m *mapIter[K, V]) Current() pair[K, V] {
	return pair[K, V]{
		Key: m.iter.Key().Interface().(K),
		Val: m.iter.Value().Interface().(V),
	}
}

type chanIter[V any] struct {
	ch <-chan V
	v  V
}

func (c *chanIter[V]) MoveNext() (ok bool) {
	c.v, ok = <-c.ch
	return
}

func (c *chanIter[V]) Current() pair[V, any] {
	return pair[V, any]{Key: c.v}
}
