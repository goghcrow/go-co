## What is go-co

go-co(routine) is a **Source to Source Compiler** which rewrites trival yield expression to monadic sequence implementation.

Inspired by [wind-js](https://github.com/JeffreyZhao/wind).

## Example

````golang
package src

import (
	"testing"

	. "github.com/goghcrow/go-co"
)

type Node[V any] struct {
	Val         V
	Left, Right *Node[V]
}

type WalkMode int

const (
	PreOrder  WalkMode = 0
	InOrder            = 1
	PostOrder          = 2
)

func Walk[V any](n *Node[V], mode WalkMode) (_ Iter[V]) {
	if n == nil {
		return
	}
	if mode == PreOrder {
		Yield(n.Val)
		YieldFrom(Walk(n.Left, mode))
		YieldFrom(Walk(n.Right, mode))
	} else if mode == InOrder {
		YieldFrom(Walk(n.Left, mode))
		Yield(n.Val)
		YieldFrom(Walk(n.Right, mode))
	} else if mode == PostOrder {
		YieldFrom(Walk(n.Left, mode))
		YieldFrom(Walk(n.Right, mode))
		Yield(n.Val)
	} else {
		panic("unknown walk mode")
	}
	return
}

// Match the same iterating path
func Match[V comparable](rootA, rootB *Node[V], mode WalkMode) bool {
	if rootA == rootB {
		return true
	}
	if rootA == nil || rootB == nil {
		return false
	}

	a, b := Walk(rootA, mode), Walk(rootB, mode)

	for {
		na, nb := a.MoveNext(), b.MoveNext()
		if na != nb {
			return false
		}
		if !na {
			return true
		}
		if a.Current() != b.Current() {
			return false
		}
	}
}

func TestTreeWalker(t *testing.T) {
	//       5
	//      / \
	//     3   7
	//    / \   \
	//   1   4   9
	//          /
	//         8
	root := &Node[int]{
		Val: 5,
		Left: &Node[int]{
			Val: 3,
			Left: &Node[int]{
				Val: 1,
			},
			Right: &Node[int]{
				Val: 4,
			},
		},
		Right: &Node[int]{
			Val: 7,
			Right: &Node[int]{
				Val: 9,
				Left: &Node[int]{
					Val: 8,
				},
			},
		},
	}

	preorder := iter2slice(Walk(root, PreOrder))
	assertEqual(t, preorder, []int{5, 3, 1, 4, 7, 9, 8})

	inorder := iter2slice(Walk(root, InOrder))
	assertEqual(t, inorder, []int{1, 3, 4, 5, 7, 8, 9})

	postorder := iter2slice(Walk(root, PostOrder))
	assertEqual(t, postorder, []int{1, 4, 3, 8, 9, 7, 5})
}

func TestTreeMatcher(t *testing.T) {
	//       1
	//      / \
	//     2   3
	rootA := &Node[int]{
		Val: 1,
		Left: &Node[int]{
			Val: 2,
		},
		Right: &Node[int]{
			Val: 3,
		},
	}

	//       3
	//      /
	//     1
	//    /
	//   2
	rootB := &Node[int]{
		Val: 3,
		Left: &Node[int]{
			Val: 1,
			Left: &Node[int]{
				Val: 2,
			},
		},
	}

	assertEqual(t, Match(rootA, rootB, PreOrder), false)
	assertEqual(t, Match(rootA, rootB, InOrder), true)
	assertEqual(t, Match(rootA, rootB, PostOrder), false)
}
````

compile output:

```golang
// Code generated by github.com/goghcrow/go-co DO NOT EDIT.
package src

import (
	ʂɘʠ "github.com/goghcrow/go-co/seq"
	"testing"
)

type Node[V any] struct {
	Val         V
	Left, Right *Node[V]
}

type WalkMode int

const (
	PreOrder  WalkMode = 0
	InOrder            = 1
	PostOrder          = 2
)

func Walk[V any](n *Node[V], mode WalkMode) (_ ʂɘʠ.Iterator[V]) {
	return ʂɘʠ.Start[V](ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
		if n == nil {
			return ʂɘʠ.Return[V]()

		}
		return ʂɘʠ.Combine[V](ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
			if mode == PreOrder {
				return ʂɘʠ.Bind[V](n.Val, func() ʂɘʠ.Seq[V] {
					return ʂɘʠ.Combine[V](
						ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
							ɪʇ := Walk(n.Left, mode)
							return ʂɘʠ.While[V](func() bool {
								return ɪʇ.MoveNext()
							}, ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
								ʌ := ɪʇ.Current()
								return ʂɘʠ.Bind[V](ʌ,
									ʂɘʠ.Normal[V],
								)
							}))
						}),

						ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
							ɪʇ := Walk(n.Right, mode)
							return ʂɘʠ.While[V](func() bool {
								return ɪʇ.MoveNext()
							}, ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
								ʌ := ɪʇ.Current()
								return ʂɘʠ.Bind[V](ʌ,
									ʂɘʠ.Normal[V],
								)
							}))
						}),
					)
				})
			} else if mode == InOrder {
				return ʂɘʠ.Combine[V](
					ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
						ɪʇ := Walk(n.Left, mode)
						return ʂɘʠ.While[V](func() bool {
							return ɪʇ.MoveNext()
						}, ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
							ʌ := ɪʇ.Current()
							return ʂɘʠ.Bind[V](ʌ,
								ʂɘʠ.Normal[V],
							)
						}))
					}),
					ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
						return ʂɘʠ.Bind[V](n.Val, func() ʂɘʠ.Seq[V] {
							return ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
								ɪʇ := Walk(n.Right, mode)
								return ʂɘʠ.While[V](func() bool {
									return ɪʇ.MoveNext()
								}, ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
									ʌ := ɪʇ.Current()
									return ʂɘʠ.Bind[V](ʌ,
										ʂɘʠ.Normal[V],
									)
								}))
							})
						})
					}))
			} else if mode == PostOrder {
				return ʂɘʠ.Combine[V](
					ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
						ɪʇ := Walk(n.Left, mode)
						return ʂɘʠ.While[V](func() bool {
							return ɪʇ.MoveNext()
						}, ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
							ʌ := ɪʇ.Current()
							return ʂɘʠ.Bind[V](ʌ,
								ʂɘʠ.Normal[V],
							)
						}))
					}),

					ʂɘʠ.Combine[V](
						ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
							ɪʇ := Walk(n.Right, mode)
							return ʂɘʠ.While[V](func() bool {
								return ɪʇ.MoveNext()
							}, ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
								ʌ := ɪʇ.Current()
								return ʂɘʠ.Bind[V](ʌ,
									ʂɘʠ.Normal[V],
								)
							}))
						}),
						ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
							return ʂɘʠ.Bind[V](n.Val,
								ʂɘʠ.Normal[V],
							)
						})),
				)
			} else {

				panic("unknown walk mode")
			}
			return ʂɘʠ.Normal[V]()
		}), ʂɘʠ.Delay[V](func() ʂɘʠ.Seq[V] {
			return ʂɘʠ.Return[V]()
		}))
	}))

}

// Match the same iterating path
func Match[V comparable](rootA, rootB *Node[V], mode WalkMode) bool {
	if rootA == rootB {
		return true
	}
	if rootA == nil || rootB == nil {
		return false
	}

	a, b := Walk(rootA, mode), Walk(rootB, mode)

	for {
		na, nb := a.MoveNext(), b.MoveNext()
		if na != nb {
			return false
		}
		if !na {
			return true
		}
		if a.Current() != b.Current() {
			return false
		}
	}
}

func TestTreeWalker(t *testing.T) {
	//       5
	//      / \
	//     3   7
	//    / \   \
	//   1   4   9
	//          /
	//         8
	root := &Node[int]{
		Val: 5,
		Left: &Node[int]{
			Val: 3,
			Left: &Node[int]{
				Val: 1,
			},
			Right: &Node[int]{
				Val: 4,
			},
		},
		Right: &Node[int]{
			Val: 7,
			Right: &Node[int]{
				Val: 9,
				Left: &Node[int]{
					Val: 8,
				},
			},
		},
	}

	preorder := iter2slice(Walk(root, PreOrder))
	assertEqual(t, preorder, []int{5, 3, 1, 4, 7, 9, 8})

	inorder := iter2slice(Walk(root, InOrder))
	assertEqual(t, inorder, []int{1, 3, 4, 5, 7, 8, 9})

	postorder := iter2slice(Walk(root, PostOrder))
	assertEqual(t, postorder, []int{1, 4, 3, 8, 9, 7, 5})
}

func TestTreeMatcher(t *testing.T) {
	//       1
	//      / \
	//     2   3
	rootA := &Node[int]{
		Val: 1,
		Left: &Node[int]{
			Val: 2,
		},
		Right: &Node[int]{
			Val: 3,
		},
	}

	//       3
	//      /
	//     1
	//    /
	//   2
	rootB := &Node[int]{
		Val: 3,
		Left: &Node[int]{
			Val: 1,
			Left: &Node[int]{
				Val: 2,
			},
		},
	}

	assertEqual(t, Match(rootA, rootB, PreOrder), false)
	assertEqual(t, Match(rootA, rootB, InOrder), true)
	assertEqual(t, Match(rootA, rootB, PostOrder), false)
}

```

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