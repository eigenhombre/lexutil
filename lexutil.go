package lexutil

// Adapted from:
// https://www.youtube.com/watch?v=HxaD_trXwRE
// https://talks.golang.org/2011/lex.slide
//
// Slightly modified by John Jacobsen to work with a user-supplied set of
// lexemes and state functions.
//
// Copyright (c) 2011 The Go Authors. All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Use of this source code is governed by a BSD-style
// license that can be found in the Go language LICENSE file.

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ItemType distinguishes between different lexemes.
type ItemType int

// LexItem represents a token returned from the lexer.
type LexItem struct {
	Typ ItemType
	Val string
}

// StateFn represents the state of the lexer; each invocation returns the next
// state.
type StateFn func(*Lexer) StateFn

// Lexer holds the current state of the lexer.
type Lexer struct {
	name  string       // used only for error reports.
	input string       // the string being scanned.
	start int          // start position of this item.
	pos   int          // current position in the input.
	width int          // width of last rune read from input.
	items chan LexItem // channel of scanned items.
}

// Emit passes a lexeme to the consumer on the items channel, and resets the
// start to the current position.
func (l *Lexer) Emit(t ItemType) {
	l.items <- LexItem{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

const eof = -1

// Next returns the next rune in the input, accounting for variable rune width
// in bytes.
func (l *Lexer) Next() rune {
	var r rune

	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// Lex starts the lexer on a separate goroutine.
func Lex(name, input string, startState StateFn) *Lexer {
	l := &Lexer{
		name:  name,
		input: input,
		items: make(chan LexItem),
	}
	go l.run(startState) // Concurrently run state machine.
	return l
}

// Peek returns but does not consume the next rune in the input.
func (l *Lexer) Peek() rune {
	r := l.Next()
	l.Backup()
	return r
}

// Backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) Backup() {
	l.pos -= l.width
}

func (l *Lexer) run(startState StateFn) {
	for state := startState; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// Ignore skips over the pending input before this point.
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Accept consumes the next rune if it's from the valid set.
func (l *Lexer) Accept(valid string) bool {
	if strings.ContainsRune(valid, l.Next()) {
		return true
	}
	l.Backup()
	return false
}

// AcceptRun consumes a run of runes from the valid set.
func (l *Lexer) AcceptRun(valid string) {
	for strings.ContainsRune(valid, l.Next()) {
	}
	l.Backup()
}

// Errorf returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *Lexer) Errorf(format string, errItem ItemType, args ...interface{}) StateFn {
	l.items <- LexItem{
		errItem,
		fmt.Sprintf(format, args...),
	}
	return nil
}
