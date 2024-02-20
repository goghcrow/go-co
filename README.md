# 𝐖𝐡𝐚𝐭 𝐢𝐬 𝕘𝕠-𝕔𝕠 

> **𝙏𝙝𝙚 𝙊𝙡𝙙 𝙉𝙚𝙬 𝙏𝙝𝙞𝙣𝙜**

golang co-routine is a source to source compiler which rewrites trival yield statements to monadic style code.

Inspired by [wind-js](https://github.com/JeffreyZhao/wind).


## Quick Start

Create source files ending with `_co.go` / `_co_test.go`.

Build tag `//go:build co` required.

Then `go generate -tags co ./...` (or run by IDE whatever).

And it is a good idea to switch custom build tag to `co` when working in goland or vscode,
so IDE will be happy to index and check your code.

```golang
//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen
//go:generate cogen

package main

import (
	. "github.com/goghcrow/go-co"
)

func Fibonacci() Iter[int] {
  a, b := 1, 1
  for {
    Yield(b)
    a, b = b, a+b
  }
}

func main() {
	for n := range Fibonacci() {
		if n > 1000 {
			println(n)
			break
		}
	}
}
```


## Example

- [Simple](example/example_co.go)
- [Tree](example/tree/tree_co.go)
- [Linq](example/linq/linq_co.go)
- [MicroThread](example/microthread/soldier_co.go)
- [Lexer](example/lexer/lexer_co.go)
- [Sched1](example/sched1/sched_co.go)
- [Sched2](example/sched2/sched_co.go)


## API

`go get github.com/goghcrow/go-co@latest`

```golang
package main

import (
    "github.com/goghcrow/go-co/rewriter"
    "github.com/goghcrow/go-loader"
)

func main() {
    rewriter.Compile(
        "./src",
        "./out",
        loader.WithLoadTest(),
    )
}
```

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
