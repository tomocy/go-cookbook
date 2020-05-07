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
			},
		},
		"string": {
			src: `"aiueo01234"`,
			expected: []token{
				{kind: tokenString, literal: `"aiueo01234"`, pos: pos{line: 0, start: 0, end: 12}},
			},
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			lex := newLexer([]rune(test.src))

			for _, expected := range test.expected {
				actual := lex.readToken()
				if err := assertToken(actual, expected); err != nil {
					t.Errorf("should have read: %s", err)
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
