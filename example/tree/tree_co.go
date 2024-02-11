//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen@main
//go:generate cogen
package tree

import (
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
	switch mode {
	case PreOrder:
		Yield(n.Val)
		YieldFrom(Walk(n.Left, mode))
		YieldFrom(Walk(n.Right, mode))
	case InOrder:
		YieldFrom(Walk(n.Left, mode))
		Yield(n.Val)
		YieldFrom(Walk(n.Right, mode))
	case PostOrder:
		YieldFrom(Walk(n.Left, mode))
		YieldFrom(Walk(n.Right, mode))
		Yield(n.Val)
	default:
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
