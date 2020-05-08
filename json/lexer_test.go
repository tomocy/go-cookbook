package json

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
				{kind: tokenEOF, pos: pos{line: 0, start: 0, end: 1}},
			},
		},
		"multiline empty": {
			src: "\n \n",
			expected: []token{
				{kind: tokenEOF, pos: pos{line: 2, start: 0, end: 1}},
			},
		},
		"number": {
			src: "1",
			expected: []token{
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenEOF, pos: pos{line: 0, start: 1, end: 2}},
			},
		},
		"string": {
			src: `"aiueo01234"`,
			expected: []token{
				{kind: tokenString, literal: `"aiueo01234"`, pos: pos{line: 0, start: 0, end: 12}},
				{kind: tokenEOF, pos: pos{line: 0, start: 12, end: 13}},
			},
		},
		"empty array": {
			src: "[]",
			expected: []token{
				{kind: tokenLBracket, literal: "[", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenRBracket, literal: "]", pos: pos{line: 0, start: 1, end: 2}},
				{kind: tokenEOF, pos: pos{line: 0, start: 2, end: 3}},
			},
		},
		"array": {
			src: `[1, "two", 3, "four", {"a": 1}]`,
			expected: []token{
				{kind: tokenLBracket, literal: "[", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 1, end: 2}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 2, end: 3}},
				{kind: tokenString, literal: `"two"`, pos: pos{line: 0, start: 4, end: 9}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 9, end: 10}},
				{kind: tokenNum, literal: "3", pos: pos{line: 0, start: 11, end: 12}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 12, end: 13}},
				{kind: tokenString, literal: `"four"`, pos: pos{line: 0, start: 14, end: 20}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 20, end: 21}},
				{kind: tokenLBrace, literal: "{", pos: pos{line: 0, start: 22, end: 23}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 0, start: 23, end: 26}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 26, end: 27}},
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 28, end: 29}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 0, start: 29, end: 30}},
				{kind: tokenRBracket, literal: "]", pos: pos{line: 0, start: 30, end: 31}},
				{kind: tokenEOF, pos: pos{line: 0, start: 31, end: 32}},
			},
		},
		"mulitline array": {
			src: `[
1,
"two",
3,
"four",
{"a": 1}
]`,
			expected: []token{
				{kind: tokenLBracket, literal: "[", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenNum, literal: "1", pos: pos{line: 1, start: 0, end: 1}},
				{kind: tokenComma, literal: ",", pos: pos{line: 1, start: 1, end: 2}},
				{kind: tokenString, literal: `"two"`, pos: pos{line: 2, start: 0, end: 5}},
				{kind: tokenComma, literal: ",", pos: pos{line: 2, start: 5, end: 6}},
				{kind: tokenNum, literal: "3", pos: pos{line: 3, start: 0, end: 1}},
				{kind: tokenComma, literal: ",", pos: pos{line: 3, start: 1, end: 2}},
				{kind: tokenString, literal: `"four"`, pos: pos{line: 4, start: 0, end: 6}},
				{kind: tokenComma, literal: ",", pos: pos{line: 4, start: 6, end: 7}},
				{kind: tokenLBrace, literal: "{", pos: pos{line: 5, start: 0, end: 1}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 5, start: 1, end: 4}},
				{kind: tokenColon, literal: ":", pos: pos{line: 5, start: 4, end: 5}},
				{kind: tokenNum, literal: "1", pos: pos{line: 5, start: 6, end: 7}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 5, start: 7, end: 8}},
				{kind: tokenRBracket, literal: "]", pos: pos{line: 6, start: 0, end: 1}},
				{kind: tokenEOF, pos: pos{line: 6, start: 1, end: 2}},
			},
		},
		"empty object": {
			src: "{}",
			expected: []token{
				{kind: tokenLBrace, literal: "{", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 0, start: 1, end: 2}},
				{kind: tokenEOF, pos: pos{line: 0, start: 2, end: 3}},
			},
		},
		"object": {
			src: `{"a": 1, "b": "two", "c": 3, "d": "four", "e": [5, "six"], "f": {"a": 1}}`,
			expected: []token{
				{kind: tokenLBrace, literal: "{", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 0, start: 1, end: 4}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 4, end: 5}},
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 6, end: 7}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 7, end: 8}},
				{kind: tokenString, literal: `"b"`, pos: pos{line: 0, start: 9, end: 12}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 12, end: 13}},
				{kind: tokenString, literal: `"two"`, pos: pos{line: 0, start: 14, end: 19}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 19, end: 20}},
				{kind: tokenString, literal: `"c"`, pos: pos{line: 0, start: 21, end: 24}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 24, end: 25}},
				{kind: tokenNum, literal: "3", pos: pos{line: 0, start: 26, end: 27}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 27, end: 28}},
				{kind: tokenString, literal: `"d"`, pos: pos{line: 0, start: 29, end: 32}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 32, end: 33}},
				{kind: tokenString, literal: `"four"`, pos: pos{line: 0, start: 34, end: 40}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 40, end: 41}},
				{kind: tokenString, literal: `"e"`, pos: pos{line: 0, start: 42, end: 45}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 45, end: 46}},
				{kind: tokenLBracket, literal: "[", pos: pos{line: 0, start: 47, end: 48}},
				{kind: tokenNum, literal: "5", pos: pos{line: 0, start: 48, end: 49}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 49, end: 50}},
				{kind: tokenString, literal: `"six"`, pos: pos{line: 0, start: 51, end: 56}},
				{kind: tokenRBracket, literal: "]", pos: pos{line: 0, start: 56, end: 57}},
				{kind: tokenComma, literal: ",", pos: pos{line: 0, start: 57, end: 58}},
				{kind: tokenString, literal: `"f"`, pos: pos{line: 0, start: 59, end: 62}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 62, end: 63}},
				{kind: tokenLBrace, literal: "{", pos: pos{line: 0, start: 64, end: 65}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 0, start: 65, end: 68}},
				{kind: tokenColon, literal: ":", pos: pos{line: 0, start: 68, end: 69}},
				{kind: tokenNum, literal: "1", pos: pos{line: 0, start: 70, end: 71}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 0, start: 71, end: 72}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 0, start: 72, end: 73}},
				{kind: tokenEOF, pos: pos{line: 0, start: 73, end: 74}},
			},
		},
		"multiline object": {
			src: `{
"a": 1,
"b": "two",
"c": 3,
"d": "four",
"e": [5, "six"],
"f": {"a": 1}
}`,
			expected: []token{
				{kind: tokenLBrace, literal: "{", pos: pos{line: 0, start: 0, end: 1}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 1, start: 0, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 1, start: 3, end: 4}},
				{kind: tokenNum, literal: "1", pos: pos{line: 1, start: 5, end: 6}},
				{kind: tokenComma, literal: ",", pos: pos{line: 1, start: 6, end: 7}},
				{kind: tokenString, literal: `"b"`, pos: pos{line: 2, start: 0, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 2, start: 3, end: 4}},
				{kind: tokenString, literal: `"two"`, pos: pos{line: 2, start: 5, end: 10}},
				{kind: tokenComma, literal: ",", pos: pos{line: 2, start: 10, end: 11}},
				{kind: tokenString, literal: `"c"`, pos: pos{line: 3, start: 0, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 3, start: 3, end: 4}},
				{kind: tokenNum, literal: "3", pos: pos{line: 3, start: 5, end: 6}},
				{kind: tokenComma, literal: ",", pos: pos{line: 3, start: 6, end: 7}},
				{kind: tokenString, literal: `"d"`, pos: pos{line: 4, start: 0, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 4, start: 3, end: 4}},
				{kind: tokenString, literal: `"four"`, pos: pos{line: 4, start: 5, end: 11}},
				{kind: tokenComma, literal: ",", pos: pos{line: 4, start: 11, end: 12}},
				{kind: tokenString, literal: `"e"`, pos: pos{line: 5, start: 0, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 5, start: 3, end: 4}},
				{kind: tokenLBracket, literal: "[", pos: pos{line: 5, start: 5, end: 6}},
				{kind: tokenNum, literal: "5", pos: pos{line: 5, start: 6, end: 7}},
				{kind: tokenComma, literal: ",", pos: pos{line: 5, start: 7, end: 8}},
				{kind: tokenString, literal: `"six"`, pos: pos{line: 5, start: 9, end: 14}},
				{kind: tokenRBracket, literal: "]", pos: pos{line: 5, start: 14, end: 15}},
				{kind: tokenComma, literal: ",", pos: pos{line: 5, start: 15, end: 16}},
				{kind: tokenString, literal: `"f"`, pos: pos{line: 6, start: 0, end: 3}},
				{kind: tokenColon, literal: ":", pos: pos{line: 6, start: 3, end: 4}},
				{kind: tokenLBrace, literal: "{", pos: pos{line: 6, start: 5, end: 6}},
				{kind: tokenString, literal: `"a"`, pos: pos{line: 6, start: 6, end: 9}},
				{kind: tokenColon, literal: ":", pos: pos{line: 6, start: 9, end: 10}},
				{kind: tokenNum, literal: "1", pos: pos{line: 6, start: 11, end: 12}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 6, start: 12, end: 13}},
				{kind: tokenRBrace, literal: "}", pos: pos{line: 7, start: 0, end: 1}},
				{kind: tokenEOF, pos: pos{line: 7, start: 1, end: 2}},
			},
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			lex := newLexer([]rune(test.src))

			for _, expected := range test.expected {
				actual := lex.readToken()
				if err := assertToken(actual, expected); err != nil {
					t.Errorf("should have read: unexpected token: %s", err)
				}
			}
		})
	}
}

func TestReadChar(t *testing.T) {
	src := "aaaaa"
	expected := []struct {
		currPos int
		nextPos int
	}{
		{currPos: 0, nextPos: 1},
		{currPos: 1, nextPos: 2},
		{currPos: 2, nextPos: 3},
		{currPos: 3, nextPos: 4},
		{currPos: 4, nextPos: 5},
		{currPos: 5, nextPos: 6},
	}

	lex := lexer{
		src: []rune(src),
	}

	for _, expected := range expected {
		lex.readChar()
		if lex.currIndex != expected.currPos {
			t.Errorf("should have read char: %s", reportUnexpected("currPos", lex.currIndex, expected.currPos))
			return
		}
		if lex.nextIndex != expected.nextPos {
			t.Errorf("should have read char: %s", reportUnexpected("nextPos", lex.nextIndex, expected.nextPos))
			return
		}
	}

	lex.readChar()
	lex.readChar()
	lex.readChar()
	lastExpected := expected[len(expected)-1]
	if lex.currIndex != lastExpected.currPos {
		t.Errorf("should have read char: %s", reportUnexpected("currPos", lex.currIndex, lastExpected.currPos))
		return
	}
	if lex.nextIndex != lastExpected.nextPos {
		t.Errorf("should have read char: %s", reportUnexpected("nextPos", lex.nextIndex, lastExpected.nextPos))
		return
	}
}

func assertToken(actual, expected token) error {
	if actual.kind != expected.kind {
		return reportUnexpected("kind", actual.kind, expected.kind)
	}
	if actual.literal != expected.literal {
		return reportUnexpected("literal", actual.literal, expected.literal)
	}
	if err := assertPos(actual.pos, expected.pos); err != nil {
		return fmt.Errorf("unexpected pos: %w", err)
	}

	return nil
}

func assertPos(actual, expected pos) error {
	if actual.line != expected.line {
		return reportUnexpected("line", actual.line, expected.line)
	}
	if actual.start != expected.start {
		return reportUnexpected("start", actual.start, expected.start)
	}
	if actual.end != expected.end {
		return reportUnexpected("end", actual.end, expected.end)
	}

	return nil
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, expected %v", name, actual, expected)
}
