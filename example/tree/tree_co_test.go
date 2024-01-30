//go:build co

package tree

import (
	"reflect"
	"testing"

	. "github.com/goghcrow/go-co"
)

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
		assertEqual(t, Match(rootA, rootB, PreOrder), false)
		assertEqual(t, Match(rootA, rootB, InOrder), true)
		assertEqual(t, Match(rootA, rootB, PostOrder), false)
	}
}

func iter2slice[V any](g Iter[V]) (xs []V) {
	for i := range g {
		xs = append(xs, i)
	}
	return xs
}

func assertEqual(t *testing.T, got, expect any) {
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect\n%+v\n\ngot\n%+v", expect, got)
	}
}
