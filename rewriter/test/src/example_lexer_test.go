package src

import (
	"reflect"
	"testing"
	"unicode"
	"unicode/utf8"

	. "github.com/goghcrow/go-co"
)

func TestLexer(t *testing.T) {
	var (
		Num TokTyp = 1
		Sym TokTyp = 2

		lexWS, lexNum, lexStr StateFn
	)
	lexWS = func(l *L) (_ Iter[Tok]) {
		if l.Peek() == EOF {
			return
		}
		if l.Consume(unicode.IsSpace) {
			l.Skip()
		}
		YieldFrom(lexNum(l))
		return
	}
	lexNum = func(l *L) (_ Iter[Tok]) {
		if l.Consume(unicode.IsDigit) {
			if l.Peek() == '.' {
				l.Nxt()
				if !l.Consume(unicode.IsDigit) {
					panic("invalid num")
				}
				Yield(l.Tok(Num))
			} else {
				Yield(l.Tok(Num))
			}
		}
		YieldFrom(lexStr(l))
		return
	}
	lexStr = func(l *L) (_ Iter[Tok]) {
		if l.Consume(unicode.IsLetter) {
			Yield(l.Tok(Sym))
		}
		YieldFrom(lexWS(l))
		return
	}

	all := []struct {
		src    string
		expect []Tok
	}{
		{src: "", expect: nil},
		{src: "  ", expect: nil},
		{src: "3.14", expect: []Tok{{Num, "3.14"}}},
		{src: " 100 ", expect: []Tok{{Num, "100"}}},
		{src: "hello", expect: []Tok{{Sym, "hello"}}},
		{src: " world ", expect: []Tok{{Sym, "world"}}},
		{
			src: "3.14  hello   world  100",
			expect: []Tok{
				{Num, "3.14"},
				{Sym, "hello"},
				{Sym, "world"},
				{Num, "100"},
			},
		},
	}

	lexer := MkLexer(lexWS)

	for _, it := range all {
		t.Run(it.src, func(t *testing.T) {
			var got []Tok
			for tok := range lexer(it.src) {
				t.Logf("[%d](%s) ", tok.Typ, tok.Val)
				got = append(got, tok)
			}
			if !reflect.DeepEqual(it.expect, got) {
				t.Errorf("\nexpect\n%+v\n\ngot\n%+v", it.expect, got)
			}
		})
	}
}

// ====================================================

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
