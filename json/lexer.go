package json

func newLexer(src []rune) lexer {
	return lexer{
		src: src,
	}
}

type lexer struct {
	src       []rune
	currIndex int
	nextIndex int
	pos       pos
}

const (
	charEOF = 0

	literalTrue  = "true"
	literalFalse = "false"
)

func (l *lexer) readToken() token {
	l.readChar()
	l.skipWhitespaces()

	switch char := l.currChar(); char {
	case charEOF:
		return token{
			kind:    tokenKinds[string(char)],
			literal: "",
			pos:     l.pos,
		}
	case '[', ']', '{', '}', ',', ':':
		return token{
			kind:    tokenKinds[string(char)],
			literal: string(char),
			pos:     l.pos,
		}
	case '"':
		return l.composeString()
	default:
		if isNum(char) {
			return l.composeNum()
		}
		if isLetter(char) {
			return l.composeLetters()
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

func (l *lexer) composeString() token {
	t := token{
		kind: tokenString,
		pos: pos{
			line:  l.pos.line,
			start: l.pos.start,
		},
	}
	t.literal = l.readString()
	t.pos.end = l.pos.end

	return t
}

func (l *lexer) readString() string {
	start := l.currIndex
	for {
		l.readChar()

		if l.currChar() == charEOF || l.currChar() == '"' {
			break
		}
	}

	return string(l.src[start:l.nextIndex])
}

func (l *lexer) composeNum() token {
	t := token{
		kind: tokenNum,
		pos: pos{
			line:  l.pos.line,
			start: l.pos.start,
		},
	}
	t.literal = l.readNumber()
	t.pos.end = l.pos.end

	return t
}

func (l *lexer) readNumber() string {
	start := l.currIndex

	for isNum(l.nextChar()) {
		l.readChar()
	}

	return string(l.src[start:l.nextIndex])
}

func (l *lexer) composeLetters() token {
	t := token{
		pos: pos{
			line:  l.pos.line,
			start: l.pos.start,
		},
	}
	t.literal = l.readLetters()
	t.pos.end = l.pos.end

	kind, ok := tokenKinds[t.literal]
	t.kind = kind
	if !ok {
		t.kind = tokenIllegal
	}

	return t
}

func (l *lexer) readLetters() string {
	start := l.currIndex

	for isLetter(l.nextChar()) {
		l.readChar()
	}

	return string(l.src[start:l.nextIndex])
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

func isLetter(c rune) bool {
	return 'a' <= c && c <= 'z'
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

	tokenLBracket tokenKind = "["
	tokenRBracket tokenKind = "]"
	tokenLBrace   tokenKind = "{"
	tokenRBrace   tokenKind = "}"
	tokenComma    tokenKind = ","
	tokenColon    tokenKind = ":"

	tokenNum    tokenKind = "number"
	tokenString tokenKind = "string"
	tokenBool   tokenKind = "bool"
)

var tokenKinds = map[string]tokenKind{
	string(charEOF): tokenEOF,
	"[":             tokenLBracket,
	"]":             tokenRBracket,
	"{":             tokenLBrace,
	"}":             tokenRBrace,
	",":             tokenComma,
	":":             tokenColon,
	"true":          tokenBool,
	"false":         tokenBool,
}

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
