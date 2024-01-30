# What is go-co

> **The Old New Thing**

go-co(routine) is a **Source to Source Compiler** which rewrites trival yield expression to monadic style code.

Inspired by [wind-js](https://github.com/JeffreyZhao/wind).

## [WIP]Control Flow Support

Rewrite control flow to monadic func invoking.

- [x] IfStmt
- [x] SwitchStmt
  - [ ] Fallthrough
- [x] TypeSwitchStmt
- [x] ForStmt
- [x] RangeStmt
  - [x] string
  - [x] slice
  - [x] map
  - [x] array
  - [x] integer
  - [x] channel
  - [ ] range func
- [x] BlockStmt
- [x] Break / Continue 
  - [x] Non-Label
  - [ ] Label
- [ ] Goto
- [ ] SelectStmt
- [ ] DeferStmt


## Quick Start

### Usage

Create source files ending with **_co.go**,  **_co_test.go**, and **build tag** required.

Run `go generate -tags co ./...` (or IDE whatever)

```golang
//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen
//go:generate cogen

package pkg

import (
	. "github.com/goghcrow/go-co"
    // or
    // "github.com/goghcrow/go-co"
)

func Fibonacci() Iter[int] {
  a, b := 1, 1
  for {
    Yield(b)
    a, b = b, a+b
  }
}
```


#### Example

- [Simple](example/example_co.go)
- [Tree](example/tree/tree_co.go)
- [Lexer](example/lexer/lexer_co.go)
- [Sched](example/sched/sched_co.go)


### API

`go get github.com/goghcrow/go-co@latest`

```golang
package main

import (
    "github.com/goghcrow/go-co/rewriter"
    "github.com/goghcrow/go-ast-matcher"
)

func main() {
    rewriter.Compile(
        "./src",
        "./out",
        matcher.PatternAll,
        matcher.WithLoadTest(),
    )
}
```


## Example

### Sample

```golang
package sample

import (
	"strings"
	
	. "github.com/goghcrow/go-co"
)

func SampleGetNumList() (_ Iter[int]) {
	Yield(1)
	Yield(2)
	Yield(3)
	return
}

func SampleYieldFrom() (_ Iter[int]) {
	Yield(0)
	YieldFrom(SampleGetNumList())
	Yield(4)
	return
}

type Pair[K, V any] struct {
    Key K
    Val V
}

func SampleLoop() (_ Iter[any]) {
    for i := 0; i < 5; i++ {
      Yield(i)
    }
  
    xs := []string{"a", "b", "c"}
    for i, n := range xs {
      Yield(Pair[int, string]{i, n})
    }
  
    for i, c := range "Hello World!" {
      Yield(Pair[int, rune]{i, c})
    }
  
    m := map[string]int{"a": 1, "b": 2, "c": 3}
    for k, v := range m {
      Yield(Pair[string, int]{k, v})
    }
  
    return
}


func SampleGetEvenNumbers(start, end int) (_ Iter[int]) {
	for i := start; i < end; i++ {
		if i%2 == 0 {
			Yield(i)
		}
	}
	return
}

func PowersOfTwo(exponent int) (_ Iter[int]) {
	for r, i := 1, 0; i < exponent; i++ {
		Yield(r)
		r *= 2
	}
	return
}

func Fibonacci() Iter[int] {
	a, b := 1, 1
	for {
		Yield(b)
		a, b = b, a+b
	}
}

func Range(start, end, step int) (_ Iter[int]) {
	for i := start; i <= end; i += step {
		Yield(i)
	}
	return
}

func Grep(s string, lines []string) (_ Iter[string]) {
	for _, line := range lines {
		if strings.Contains(line, s) {
			Yield(line)
		}
	}
	return
}

func Run() {
	// PUSH MODE
	for n := range SampleGetNumList() {
		println(n)
	}

	for n := range SampleYieldFrom() {
		println(n)
	}

	for n := range Fibonacci() {
		if n > 1000 {
			Yield(n)
			break
		}
	}

	for i := range Range(0, 100, 2) {
		println(i)
	}

	// or Pull Mode

	iter := Range(0, 100, 2)
	for iter.MoveNext() {
		println(iter.Current())
	}
}

```

### Tree Walker

````golang
package sample

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

### Coroutine Scheduler

```golang
package sample

import (
	"sync"
	"testing"
	"time"

	. "github.com/goghcrow/go-co"
)

var wg = &sync.WaitGroup{}

type Sched struct {
	val any
	err error
}

func (s *Sched) run(co func(s *Sched) Iter[Async]) {
	it := co(s)

	var run func()
	run = func() {
		for it.MoveNext() {
			switch v := it.Current().(type) {
			case Async:
				v.Begin(func(v any, err error) {
					s.send(v, err)
					run()
				})
				return
			default:
				panic("unreached")
			}
		}
		wg.Done()
	}

	wg.Add(1)
	run()
}

func (s *Sched) send(v any, err error) {
	s.val = v
	s.err = err
}

func (s *Sched) GetReceive() (any, error) {
	return s.val, s.err
}

func Co(co func(s *Sched) Iter[Async]) {
	sched := &Sched{}
	sched.run(co)
}

type Async interface {
	Begin(cont func(v any, err error))
}

type AsyncFun func(cont func(v any, err error))

func (f AsyncFun) Begin(cont func(v any, err error)) {
	f(cont)
}

// ------------------------------------------------------------

func Sleep(d time.Duration) Async {
	return AsyncFun(func(cont func(v any, err error)) {
		timeAfter(d, func() {
			cont(nil, nil)
		})
	})
}

func SampleAsyncTask(v any) Async {
	return AsyncFun(func(cont func(v any, err error)) {
		timeAfter(time.Second, func() {
			cont(v, nil)
		})
	})
}

func TestCo(t *testing.T) {
	Co(func(s *Sched) (_ Iter[Async]) {
		t.Log("start")

		t.Log(now() + " before sleep")
		Yield(Sleep(time.Second * 1))

		t.Log(now() + " before async task")
		Yield(SampleAsyncTask(42))

        t.Log(now() + " after async task and get result")
		result, _ := s.GetReceive()
		t.Log(result)

		t.Log("end")
		return
	})

	wg.Wait()
}

// ------------------------------------------------------------

// fake callback
func timeAfter(d time.Duration, cb func()) {
	go func() {
		time.Sleep(d)
		cb()
	}()
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
```