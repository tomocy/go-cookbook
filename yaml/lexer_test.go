package yaml

import (
	"fmt"
	"testing"
)

func TestReadToken(t *testing.T) {
	tests := map[string]struct {
		src      string
		expected []token
	}{
		"empty": {
			src: "",
			expected: []token{
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 0, end: 1}},
			},
		},
		"indent with tab": {
			src: "\t1",
			expected: []token{
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 2, end: 3}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 3, end: 4}},
			},
		},
		"number": {
			src: "1000",
			expected: []token{
				{kind: tokenNum, literal: "1000", pos: pos{line: 0, start: 0, end: 4}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 4, end: 5}},
			},
		},
		"string without quotations": {
			src: "aiueo",
			expected: []token{
				{kind: tokenString, literal: `"aiueo"`, pos: pos{line: 0, start: 0, end: 5}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 5, end: 6}},
			},
		},
		"string with quotations": {
			src: `"aiueo"`,
			expected: []token{
				{kind: tokenString, literal: `"aiueo"`, pos: pos{line: 0, start: 0, end: 7}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 7, end: 8}},
			},
		},
		"true": {
			src: "true",
			expected: []token{
				{kind: tokenBool, literal: "true", pos: pos{line: 0, start: 0, end: 4}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 4, end: 5}},
			},
		},
		"false": {
			src: "false",
			expected: []token{
				{kind: tokenBool, literal: "false", pos: pos{line: 0, start: 0, end: 5}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 0, start: 5, end: 6}},
			},
		},
		"array": {
			src: `- 1
- two
- true
- false
- a: 1`,
			expected: []token{
				{kind: tokenHyphen, literal: "-", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 2, end: 3}},
				{kind: tokenHyphen, literal: "-", pos: pos{line: 1, start: 0, end: 1}},
				{kind: tokenString, literal: `"two"`, pos: pos{line: 1, start: 2, end: 5}},
				{kind: tokenHyphen, literal: "-", pos: pos{line: 2, start: 0, end: 1}},
				{kind: tokenBool, literal: "true", pos: pos{line: 2, start: 2, end: 6}},
				{kind: tokenHyphen, literal: "-", pos: pos{line: 3, start: 0, end: 1}},
				{kind: tokenBool, literal: "false", pos: pos{line: 3, start: 2, end: 7}},
				{kind: tokenHyphen, literal: "-", pos: pos{line: 4, start: 0, end: 1}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 4, start: 2, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 4, start: 3, end: 4}},
				{kind: tokenNum, literal: "1", pos: pos{line: 4, start: 5, end: 6}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 4, start: 6, end: 7}},
			},
		},
		"dictionary": {
			src: `a: 1
b: 2
c:
  3
e:
    4`,
			expected: []token{
				{kind: tokenString, literal: `"a"`, pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 1, end: 2}},
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 3, end: 4}},
				{kind: tokenString, literal: `"b"`, pos: pos{line: 1, start: 0, end: 1}},
				{kind: tokenColon, literal: ":", pos: pos{line: 1, start: 1, end: 2}},
				{kind: tokenNum, literal: "2", pos: pos{line: 1, start: 3, end: 4}},
				{kind: tokenString, literal: `"c"`, pos: pos{line: 2, start: 0, end: 1}},
				{kind: tokenColon, literal: ":", pos: pos{line: 2, start: 1, end: 2}},
				{kind: tokenNum, literal: "3", pos: pos{line: 3, start: 2, end: 3}},
				{kind: tokenString, literal: `"e"`, pos: pos{line: 4, start: 0, end: 1}},
				{kind: tokenColon, literal: ":", pos: pos{line: 4, start: 1, end: 2}},
				{kind: tokenNum, literal: "4", pos: pos{line: 5, start: 4, end: 5}},
				{kind: tokenEOF, literal: "\x00", pos: pos{line: 5, start: 5, end: 6}},
			},
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			lex := newLexer([]rune(test.src))

			for _, expected := range test.expected {
				actual := lex.readToken()
				if err := assertToken(actual, expected); err != nil {
					t.Errorf("should have read token: %s", err)
					return
				}
			}
		})
	}
}

func assertToken(actual, expected token) error {
	if actual.kind != expected.kind {
		return reprotUnexpected("kind", actual.kind, expected.kind)
	}
	if actual.literal != expected.literal {
		return reprotUnexpected("literal", actual.literal, expected.literal)
	}
	if err := assertPos(actual.pos, expected.pos); err != nil {
		return fmt.Errorf("unexpected pos: %w", err)
	}

	return nil
}

func assertPos(actual, expected pos) error {
	if actual.line != expected.line {
		return reprotUnexpected("line", actual.line, expected.line)
	}
	if actual.start != expected.start {
		return reprotUnexpected("start", actual.start, expected.start)
	}
	if actual.end != expected.end {
		return reprotUnexpected("end", actual.end, expected.end)
	}

	return nil
}

func reprotUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
