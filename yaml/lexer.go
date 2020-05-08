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

	literalTrue  = "true"
	literalFalse = "false"
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

		return l.composeSingleToken()
	case ':':
		if l.nextChar() != ' ' && l.nextChar() != '\n' {
			return l.composeLetters()
		}

		return l.composeSingleToken()
	case '"':
		return l.composeStringWithQuotes()
	default:
		if isNum(char) {
			return l.composeNum()
		}
		if isLetter(char) {
			return l.composeLetters()
		}

		return l.composeSingleTokenAs(tokenUnknown)
	}
}

func (l *lexer) composeSingleToken() token {
	t := token{
		kind:    tokenKinds[string(l.currChar())],
		literal: string(l.currChar()),
		pos:     l.pos,
	}

	l.readChar()

	return t
}

func (l *lexer) composeSingleTokenAs(kind tokenKind) token {
	t := token{
		kind:    kind,
		literal: string(l.currChar()),
		pos:     l.pos,
	}

	l.readChar()

	return t
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
		if l.isHandlingProp() {
			break
		}

		l.readChar()
	}

	return string(l.src[start:l.currIndex])
}

func (l lexer) isHandlingProp() bool {
	return l.currChar() == ':' && l.nextChar() == ' ' || l.currChar() == ':' && l.nextChar() == '\n'
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

	l.movePos()

	l.currIndex = l.nextIndex
	l.nextIndex++
}

func (l *lexer) movePos() {
	c := l.currChar()
	if l.willReadFirstChar() {
		c = 0
	}
	l.pos.move(c)
}

func (l lexer) willReadFirstChar() bool {
	return l.currIndex == 0 && l.nextIndex == 0
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
	return ' ' <= c && c <= '~'
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

const (
	spacesInTab = 2
)

func (p *pos) move(c rune) {
	if c == '\n' {
		p.line++
		p.start, p.end = 0, 1
		return
	}
	if c == '\t' {
		for i := 0; i < spacesInTab; i++ {
			p.move(' ')
		}
		return
	}

	p.start = p.end
	p.end++
}
