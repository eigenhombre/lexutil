package lexutil

import (
	"reflect"
	"strings"
	"testing"
)

// a little language:
//
// examples: AAA BBBB   AA B AAA
// anything else fails
// demonstrate this with tests

const (
	itemAs ItemType = iota
	itemBs
)

func lexBetween(l *Lexer) StateFn {
	for {
		r := l.Next()
		if r == 'A' {
			l.AcceptRun("A")
			l.Emit(itemAs)
			return lexBetween
		}
		if r == 'B' {
			l.AcceptRun("B")
			l.Emit(itemBs)
			return lexBetween
		}
		if strings.ContainsRune(" ", r) {
			l.Ignore()
			continue
		}
		if r == EOF {
			return nil
		}
		// FIXME: itemAs
		l.Errorf("unexpected rune %q", itemAs, r)
	}
}

func items(l *Lexer) []LexItem {
	var items []LexItem
	for item := range l.Items {
		items = append(items, item)
	}
	return items
}

func TestSimpleGrammar(t *testing.T) {
	abbrev := func(typ ItemType) func(string) LexItem {
		return func(input string) LexItem {
			return LexItem{Typ: typ, Val: input}
		}
	}
	As := abbrev(itemAs)
	Bs := abbrev(itemBs)
	toks := func(items ...LexItem) []LexItem {
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
		// FIXME: Add check for error:
		// {"AA C AAA", ... etc.

	}
	for _, test := range tests {
		l := Lex("test", test.input, lexBetween)
		got := items(l)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
}
