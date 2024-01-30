//go:build co

package lexer

import (
	"reflect"
	"testing"
	"unicode"

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
