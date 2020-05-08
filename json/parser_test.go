package json

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		src      string
		expected value
	}{
		"int": {
			src:      "1",
			expected: Num(1),
		},
		"string": {
			src:      `"aiueo"`,
			expected: String("aiueo"),
		},
		"true": {
			src:      "true",
			expected: Bool(true),
		},
		"false": {
			src:      "false",
			expected: Bool(false),
		},
		"empty array": {
			src:      `[]`,
			expected: Array{},
		},
		"array": {
			src: `[1, "two", 3, "four", [1, "two"], {"a": 1, "b": "two", "c": true}, false]`,
			expected: Array{
				Num(1),
				String("two"),
				Num(3),
				String("four"),
				Array{
					Num(1),
					String("two"),
				},
				Object{
					Prop{
						key: String("a"),
						val: Num(1),
					},
					Prop{
						key: String("b"),
						val: String("two"),
					},
					Prop{
						key: String("c"),
						val: Bool(true),
					},
				},
				Bool(false),
			},
		},
		"empty object": {
			src:      `{}`,
			expected: Object{},
		},
		"object": {
			src: `{"a": 1, "b": "two", "c": 3, "d": "four", "e": [1, "two", 3], "f": {"a": 1, "b": "two", "c": false}, "g": true}`,
			expected: Object{
				Prop{
					key: String("a"),
					val: Num(1),
				},
				Prop{
					key: String("b"),
					val: String("two"),
				},
				Prop{
					key: String("c"),
					val: Num(3),
				},
				Prop{
					key: String("d"),
					val: String("four"),
				},
				Prop{
					key: String("e"),
					val: Array{
						Num(1),
						String("two"),
						Num(3),
					},
				},
				Prop{
					key: String("f"),
					val: Object{
						Prop{
							key: String("a"),
							val: Num(1),
						},
						Prop{
							key: String("b"),
							val: String("two"),
						},
						Prop{
							key: String("c"),
							val: Bool(false),
						},
					},
				},
				Prop{
					key: String("g"),
					val: Bool(true),
				},
			},
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			lex := newLexer([]rune(test.src))
			p := newParser(lex)

			actual, err := p.parse()
			if err != nil {
				t.Errorf("should have parsed: %s", err)
				return
			}
			if err := assertValue(actual, test.expected); err != nil {
				t.Errorf("should have parsed: unexpected value: %s", err)
				return
			}
		})
	}
}

func assertValue(actual, expected value) error {
	switch expected := expected.(type) {
	case Array:
		if err := assertArray(actual.(Array), expected); err != nil {
			return fmt.Errorf("unexpected array: %s", err)
		}

		return nil
	case Object:
		if err := assertObject(actual.(Object), expected); err != nil {
			return fmt.Errorf("unexpected object: %s", err)
		}

		return nil
	case Num:
		if err := assertNum(actual.(Num), expected); err != nil {
			return fmt.Errorf("unexpected num: %s", err)
		}

		return nil
	case String:
		if err := assertString(actual.(String), expected); err != nil {
			return fmt.Errorf("unexpected string: %s", err)
		}

		return nil
	case Bool:
		if err := assertBool(actual.(Bool), expected); err != nil {
			return fmt.Errorf("unexpected bool: %s", err)
		}

		return nil
	default:
		return fmt.Errorf("development error: unknown type of value: %T", expected)
	}
}

func assertArray(actual, expected Array) error {
	if len(actual) != len(expected) {
		return reportUnexpected("len of value", len(actual), len(expected))
	}
	for i, expected := range expected {
		if err := assertValue(actual[i], expected); err != nil {
			return fmt.Errorf("unexpected value at %d: %s", i, err)
		}
	}

	return nil
}

func assertObject(actual, expected Object) error {
	if len(actual) != len(expected) {
		return reportUnexpected("len of value", len(actual), len(expected))
	}
	for i, expected := range expected {
		if err := assertProp(actual[i], expected); err != nil {
			return fmt.Errorf("unexpected prop at %d: %s", i, err)
		}
	}

	return nil
}

func assertProp(actual, expected Prop) error {
	if err := assertString(actual.key, expected.key); err != nil {
		return fmt.Errorf("unexpected key: %s", err)
	}
	if err := assertValue(actual.val, expected.val); err != nil {
		return fmt.Errorf("unexpected value: %s", err)
	}

	return nil
}

func assertNum(actual, expected Num) error {
	if actual != expected {
		return reportUnexpected("value", actual, expected)
	}

	return nil
}

func assertString(actual, expected String) error {
	if actual != expected {
		return reportUnexpected("value", actual, expected)
	}

	return nil
}

func assertBool(actual, expected Bool) error {
	if actual != expected {
		return reportUnexpected("value", actual, expected)
	}

	return nil
}
