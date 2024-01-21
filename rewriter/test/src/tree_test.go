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

func WalkSwitch[V any](n *Node[V], mode WalkMode) (_ Iter[V]) {
	if n == nil {
		return
	}
	switch mode {
	case PreOrder:
		Yield(n.Val)
		YieldFrom(WalkSwitch(n.Left, mode))
		YieldFrom(WalkSwitch(n.Right, mode))
	case InOrder:
		YieldFrom(WalkSwitch(n.Left, mode))
		Yield(n.Val)
		YieldFrom(WalkSwitch(n.Right, mode))
	case PostOrder:
		YieldFrom(WalkSwitch(n.Left, mode))
		YieldFrom(WalkSwitch(n.Right, mode))
		Yield(n.Val)
	default:
		panic("unknown walk mode")
	}
	return
}

// Match the same iterating path
func Match[V comparable](rootA, rootB *Node[V], mode WalkMode, Walk func(*Node[V], WalkMode) Iter[V]) bool {
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

	{
		preorder := iter2slice(Walk(root, PreOrder))
		assertEqual(t, preorder, []int{5, 3, 1, 4, 7, 9, 8})

		inorder := iter2slice(Walk(root, InOrder))
		assertEqual(t, inorder, []int{1, 3, 4, 5, 7, 8, 9})

		postorder := iter2slice(Walk(root, PostOrder))
		assertEqual(t, postorder, []int{1, 4, 3, 8, 9, 7, 5})
	}
	{
		preorder := iter2slice(WalkSwitch(root, PreOrder))
		assertEqual(t, preorder, []int{5, 3, 1, 4, 7, 9, 8})

		inorder := iter2slice(WalkSwitch(root, InOrder))
		assertEqual(t, inorder, []int{1, 3, 4, 5, 7, 8, 9})

		postorder := iter2slice(WalkSwitch(root, PostOrder))
		assertEqual(t, postorder, []int{1, 4, 3, 8, 9, 7, 5})
	}
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

	{
		assertEqual(t, Match(rootA, rootB, PreOrder, Walk[int]), false)
		assertEqual(t, Match(rootA, rootB, InOrder, Walk[int]), true)
		assertEqual(t, Match(rootA, rootB, PostOrder, Walk[int]), false)
	}
	{
		assertEqual(t, Match(rootA, rootB, PreOrder, WalkSwitch[int]), false)
		assertEqual(t, Match(rootA, rootB, InOrder, WalkSwitch[int]), true)
		assertEqual(t, Match(rootA, rootB, PostOrder, WalkSwitch[int]), false)
	}
}
