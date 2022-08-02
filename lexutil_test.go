package lexutil

import (
	"reflect"
	"strings"
	"testing"
)

// Build, and lex, a little language:
//
// well-formed examples:
//   AAA BBBB
//   AA B AAA
// failing examples:
//   AA C AAA

const (
	itemAs ItemType = iota
	itemBs
	itemErr
)

// lexStart is the only state in the lexer.
// Most lexers will have multiple StatFn's; only
// one state is needed for our toy language.
func lexStart(l *Lexer) StateFn {
	for {
		r := l.Next()
		if r == 'A' {
			l.AcceptRun("A")
			l.Emit(itemAs)
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
	As := abbrev(itemAs)
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
		{"A", toks(As("A"))},
		{"BB BB", toks(Bs("BB"), Bs("BB"))},
		{"AAA BBBB", toks(As("AAA"), Bs("BBBB"))},
		{"AA B AAA", toks(As("AA"), Bs("B"), As("AAA"))},
		{"AA C AAA", toks(As("AA"), Err("unexpected rune 'C'"), As("AAA"))},
	}
	for _, test := range tests {
		l := Lex("test", test.input, lexStart)
		got := getitems(l)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
}
