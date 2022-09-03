package lexutil

import (
	"reflect"
	"strings"
	"testing"
)

// Build, and lex, a little language, made of water, and the letter B:
//
// well-formed examples:
//   水水水 BBBB
//   水 B 水水水
// failing examples:
//   水水 C BBB
//   i fail
//   水果

const (
	item水s ItemType = iota
	itemBs
	itemErr
)

// lexStart is the only state in the lexer.
// Most lexers will have multiple StatFn's; only
// one state is needed for our toy language.
func lexStart(l *Lexer) StateFn {
	for {
		r := l.Next()
		if r == '水' {
			l.AcceptRun("水")
			l.Emit(item水s)
			return lexStart
		}
		if r == 'B' {
			l.AcceptRun("B")
			l.Emit(itemBs)
			return lexStart
		}
		if strings.ContainsRune(" ", r) {
			l.Ignore()
			continue
		}
		if r == EOF {
			return nil
		}
		l.Errorf("unexpected rune %q", itemErr, r)
	}
}

func TestSimpleGrammar(t *testing.T) {
	abbrev := func(typ ItemType) func(string) LexItem {
		return func(input string) LexItem {
			return LexItem{Typ: typ, Val: input}
		}
	}
	As := abbrev(item水s)
	Bs := abbrev(itemBs)
	Err := abbrev(itemErr)

	toks := func(items ...LexItem) []LexItem {
		return items
	}

	getitems := func(l *Lexer) []LexItem {
		var items []LexItem
		for item := range l.Items {
			items = append(items, item)
		}
		return items
	}

	var tests = []struct {
		input string
		want  []LexItem
	}{
		{"", toks()},
		{"水", toks(As("水"))},
		{"BB BB", toks(Bs("BB"), Bs("BB"))},
		{"水水水 BBBB", toks(As("水水水"), Bs("BBBB"))},
		{"水水 B 水水水", toks(As("水水"), Bs("B"), As("水水水"))},
		{"水水 C 水水水", toks(As("水水"), Err("unexpected rune 'C'"), As("水水水"))},
	}
	for _, test := range tests {
		l := Lex("test", test.input, lexStart)
		got := getitems(l)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
}
