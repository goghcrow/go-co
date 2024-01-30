# What is go-co

> **The Old New Thing**

go-co(routine) is a **Source to Source Compiler** which rewrites trival yield expression to monadic style code.

Inspired by [wind-js](https://github.com/JeffreyZhao/wind).

## Control Flow Support

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


## Example

- [Simple](example/example_co.go)
- [Tree](example/tree/tree_co.go)
- [Lexer](example/lexer/lexer_co.go)
- [Sched](example/sched/sched_co.go)


## API

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