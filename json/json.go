package json

func newLexer(src []rune) lexer {
	lex := lexer{
		src: src,
	}
	lex.readChar()

	return lex
}

type lexer struct {
	src       []rune
	currIndex int
	nextIndex int
	pos       pos
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
	case '"':
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
	default:
		if isNum(char) {
			t := token{
				kind: tokenNum,
				pos: pos{
					line:  l.pos.line,
					start: l.pos.start,
				},
			}
			t.literal = l.readNumber()
			t.pos.end = l.pos.start

			return t
		}

		return token{
			kind: tokenIllegal,
			pos:  l.pos,
		}
	}
}

func (l *lexer) skipWhitespaces() {
	for isWhitespace(l.currChar()) {
		l.readChar()
	}
}

func (l *lexer) readNumber() string {
	start := l.currIndex

	for isNum(l.currChar()) {
		l.readChar()
	}

	return string(l.src[start:l.currIndex])
}

func (l *lexer) readString() string {
	start := l.currIndex
	for {
		l.readChar()

		if l.currChar() == charEOF {
			break
		}
		if l.currChar() == '"' {
			l.readChar()
			break
		}
	}

	return string(l.src[start:l.currIndex])
}

func (l *lexer) readChar() {
	if l.nextIndex > len(l.src) {
		return
	}

	l.currIndex = l.nextIndex
	l.nextIndex++

	l.pos.move(l.currChar())
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

func isWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

type token struct {
	kind    tokenKind
	literal string
	pos     pos
}

type tokenKind string

const (
	tokenIllegal tokenKind = "illegal"
	tokenEOF     tokenKind = "EOF"

	tokenNum    tokenKind = "number"
	tokenString tokenKind = "string"
)

type pos struct {
	line       int
	start, end int
}

func (p *pos) move(c rune) {
	if c == '\n' {
		p.line++
		p.start, p.end = 0, 0
		return
	}

	p.start = p.end
	p.end++
}
