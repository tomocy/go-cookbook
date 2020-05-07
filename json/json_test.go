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
		"object": {
			src:      "{}",
			expected: object{},
		},
		"array": {
			src:      "[]",
			expected: array{},
		},
		"int": {
			src:      "1",
			expected: intVal(1),
		},
		"string": {
			src:      `"aiueo"`,
			expected: stringVal("aiueo"),
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			actual, err := parse(test.src)
			if err != nil {
				t.Errorf("should have parsed: %s", err)
				return
			}

			if err := assertValue(actual, test.expected); err != nil {
				t.Errorf("should have parsed: %s", err)
				return
			}
		})
	}
}

func assertValue(actual, expected value) error {
	switch expected := expected.(type) {
	case object:
		return assertObject(actual.(object), expected)
	case array:
		return assertArray(actual.(array), expected)
	case intVal:
		return assertInt(actual.(intVal), expected)
	case stringVal:
		return assertString(actual.(stringVal), expected)
	default:
		return fmt.Errorf("unknown type")
	}
}

func TestParseObject(t *testing.T) {
	src := `{
		"a": 1,
		"b": "bb",
		"c": {
			"d": 2
		},
		"e": []
	}`
	expected := object{
		props: []prop{
			{
				key: "a",
				val: intVal(1),
			},
			{
				key: "b",
				val: stringVal("bb"),
			},
			{
				key: "c",
				val: object{
					props: []prop{
						{
							key: "d",
							val: intVal(2),
						},
					},
				},
			},
			{
				key: "e",
				val: array{},
			},
		},
	}
	actual, err := parseObject(src)
	if err != nil {
		t.Errorf("should have parsed object: %s", err)
		return
	}

	if err := assertObject(actual, expected); err != nil {
		t.Errorf("should have parsed object: %s", err)
		return
	}
}

func assertObject(actual, expected object) error {
	if len(actual.props) != len(expected.props) {
		return reportUnexpected("len of props", len(actual.props), len(expected.props))
	}
	for i := range expected.props {
		if err := assertProp(actual.props[i], expected.props[i]); err != nil {
			return reportUnexpected(fmt.Sprintf("prop at %d", i), actual.props[i], expected.props[i])
		}
	}

	return nil
}

func TestParseProp(t *testing.T) {
	src := `"key": "value"`
	expected := prop{
		key: "key",
		val: stringVal("value"),
	}
	actual, err := parseProp(src)
	if err != nil {
		t.Errorf("should have parsed prop: %s", err)
		return
	}

	if err := assertProp(actual, expected); err != nil {
		t.Errorf("should have parsed prop: %s", err)
		return
	}
}

func assertProp(actual, expected prop) error {
	if actual.key != expected.key {
		return reportUnexpected("key", actual.key, expected.key)
	}
	if err := assertValue(actual.val, expected.val); err != nil {
		return reportUnexpected("val", actual.val, expected.val)
	}

	return nil
}

func TestParseArray(t *testing.T) {
	src := `[
		1, 
		"a", 
		2, 
		3
	]`
	expected := array{
		intVal(1),
		stringVal("a"),
		intVal(2),
		intVal(3),
	}
	actual, err := parseArray(src)
	if err != nil {
		t.Errorf("should have parsed array: %s", err)
		return
	}

	if err := assertArray(actual, expected); err != nil {
		t.Errorf("should have parsed array: %s", err)
		return
	}
}

func assertArray(actual, expected array) error {
	if len(actual) != len(expected) {
		return reportUnexpected("len of items", len(actual), len(expected))
	}
	for i := range expected {
		if actual[i] != expected[i] {
			return reportUnexpected(fmt.Sprintf("val at %d", i), actual, expected)
		}
	}

	return nil
}

func TestParseInt(t *testing.T) {
	src := `1`
	expected := intVal(1)
	actual, err := parseInt(src)
	if err != nil {
		t.Errorf("should have parsed int: %s", err)
		return
	}

	if err := assertInt(actual, expected); err != nil {
		t.Errorf("should have parsed int: %s", err)
		return
	}
}

func assertInt(actual, expected intVal) error {
	if actual != expected {
		return reportUnexpected("val", actual, expected)
	}

	return nil
}

func TestParseString(t *testing.T) {
	src := `"aiueo"`
	expected := stringVal("aiueo")
	actual, err := parseString(src)
	if err != nil {
		t.Errorf("should have parsed string: %s", err)
		return
	}

	if err := assertString(actual, expected); err != nil {
		t.Errorf("should have parsed string: %s", err)
		return
	}
}

func assertString(actual, expected stringVal) error {
	if actual != expected {
		return reportUnexpected("val", actual, expected)
	}

	return nil
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
