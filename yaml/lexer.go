package yaml

import (
	"fmt"
)

func newLexer(src []rune) lexer {
	lex := lexer{
		src: src,
	}
	lex.readChar()

	return lex
}

type lexer struct {
	src                  []rune
	currIndex, nextIndex int
	pos                  pos
}

const (
	charEOF = 0
)

func (l *lexer) readToken() token {
	l.skipWhitespaces()

	switch char := l.currChar(); char {
	case charEOF:
		return token{
			kind: tokenEOF,
			pos:  l.pos,
		}
	case '-':
		if l.nextChar() != ' ' {
			return l.composeLetters()
		}

		t := token{
			kind:    tokenKinds[string(char)],
			literal: string(char),
			pos:     l.pos,
		}
		l.readChar()

		return t
	case ':':
		if l.nextChar() != ' ' && l.nextChar() != '\n' {
			return l.composeLetters()
		}

		t := token{
			kind:    tokenKinds[string(char)],
			literal: string(char),
			pos:     l.pos,
		}
		l.readChar()

		return t
	case '"':
		return l.composeStringWithQuotes()
	default:
		if isNum(char) {
			return l.composeNum()
		}
		if isLetter(char) {
			return l.composeLetters()
		}

		return token{
			kind: tokenUnknown,
		}
	}
}

func (l *lexer) readString() string {
	start := l.currIndex
	for {
		l.readChar()
		if l.currChar() == '"' {
			l.readChar()
			break
		}
	}

	return string(l.src[start:l.currIndex])
}

func (l *lexer) composeStringWithQuotes() token {
	t := token{
		kind: tokenString,
		pos: pos{
			line:  l.pos.line,
			start: l.pos.start,
		},
	}
	t.literal = l.readString()
	t.pos.end = l.pos.start

	return t
}

func (l *lexer) composeNum() token {
	t := token{
		kind: tokenNum,
		pos: pos{
			line:  l.pos.line,
			start: l.pos.start,
		},
	}
	t.literal = l.readNum()
	t.pos.end = l.pos.start

	return t
}

func (l *lexer) readNum() string {
	start := l.currIndex
	for isNum(l.currChar()) {
		l.readChar()
	}

	return string(l.src[start:l.currIndex])
}

func (l *lexer) composeLetters() token {
	t := token{
		pos: pos{
			line:  l.pos.line,
			start: l.pos.start,
		},
	}

	lit := l.readLetters()
	t.pos.end = l.pos.start

	if kind, ok := tokenKinds[lit]; ok {
		t.kind, t.literal = kind, lit
		return t
	}

	t.kind, t.literal = tokenString, quoteString(lit)
	return t
}

func (l *lexer) readLetters() string {
	start := l.currIndex
	for isLetter(l.currChar()) {
		l.readChar()
	}

	return string(l.src[start:l.currIndex])
}

func (l *lexer) skipWhitespaces() {
	for isWhitespaces(l.currChar()) {
		l.readChar()
	}
}

func isWhitespaces(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func (l *lexer) readChar() {
	if l.nextIndex > len(l.src) {
		return
	}

	l.pos.move(l.currChar())

	l.currIndex = l.nextIndex
	l.nextIndex++
}

func (l lexer) currChar() rune {
	if l.currIndex >= len(l.src) {
		return charEOF
	}

	return l.src[l.currIndex]
}

func (l lexer) nextChar() rune {
	if l.nextIndex >= len(l.src) {
		return charEOF
	}

	return l.src[l.nextIndex]
}

func isNum(c rune) bool {
	return '0' <= c && c <= '9'
}

func isLetter(c rune) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func quoteString(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

type token struct {
	kind    tokenKind
	literal string
	pos     pos
}

type tokenKind string

const (
	tokenUnknown tokenKind = "unknown"
	tokenEOF     tokenKind = "EOF"

	tokenHyphen tokenKind = "-"
	tokenColon  tokenKind = ":"

	tokenNum    tokenKind = "number"
	tokenString tokenKind = "string"
	tokenBool   tokenKind = "bool"
)

var tokenKinds = map[string]tokenKind{
	"-":     tokenHyphen,
	":":     tokenColon,
	"true":  tokenBool,
	"false": tokenBool,
}

type pos struct {
	line       int
	start, end int
}

func (p *pos) move(c rune) {
	if c == '\n' {
		p.line++
		p.start, p.end = 0, 1
		return
	}

	p.start = p.end
	p.end++
}
