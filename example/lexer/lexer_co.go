//go:build co

//go:generate go install github.com/goghcrow/go-co/cmd/cogen
//go:generate cogen

package lexer

import (
	"unicode/utf8"

	. "github.com/goghcrow/go-co"
)

const EOF rune = -1

type (
	pos     = int
	StateFn func(*L) Iter[Tok]
	Lexer   func(string) Iter[Tok]
	TokTyp  int
	Tok     struct {
		Typ TokTyp
		Val string
	}
)

func MkLexer(fn StateFn) Lexer {
	return func(src string) (_ Iter[Tok]) {
		YieldFrom(fn(&L{src: src}))
		return
	}
}

type L struct {
	src   string
	x, y  pos // span
	stack *node
}

func (l *L) Skip() {
	l.stack = nil
	l.x = l.y
}

func (l *L) Span() string {
	return l.src[l.x:l.y]
}

func (l *L) Nxt() (r rune) {
	sz := 0
	rest := l.src[l.y:]
	if len(rest) == 0 {
		r = EOF
	} else {
		r, sz = utf8.DecodeRuneInString(rest)
	}
	l.y += sz
	l.stack = l.stack.push(r)
	return
}

func (l *L) Rewind() {
	var r rune
	r, l.stack = l.stack.pop()

	if r > EOF {
		l.y -= utf8.RuneLen(r)
		l.x = min(l.x, l.y)
	}
}

func (l *L) Peek() (r rune) {
	r = l.Nxt()
	l.Rewind()
	return
}

func (l *L) Tok(t TokTyp) (tok Tok) {
	tok = Tok{Typ: t, Val: l.Span()}
	l.x = l.y
	l.stack = nil
	return
}

func (l *L) Consume(f func(rune) bool) bool {
	r := l.Nxt()
	for f(r) {
		r = l.Nxt()
	}
	l.Rewind()
	return l.y > l.x
}

type node struct {
	pre *node
	r   rune
}

func (n *node) push(r rune) *node  { return &node{n, r} }
func (n *node) pop() (rune, *node) { return n.r, n.pre } // unsafe

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
