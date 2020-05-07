package yaml

import (
	"fmt"
	"testing"
)

func TestParseObject(t *testing.T) {
	src := `status: 200
message: "success"
resource:
  id: 10
  name: "aiueo"`
	expected := object{
		fields: []field{
			field{
				key: "status",
				val: intVal(200),
			},
			field{
				key: "message",
				val: stringVal("success"),
			},
			field{
				key: "resource",
				val: object{
					fields: []field{
						field{
							key: "id",
							val: intVal(10),
						},
						field{
							key: "name",
							val: stringVal("aiueo"),
						},
					},
				},
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
	if len(actual.fields) != len(expected.fields) {
		return reportUnexpected("len of fields", len(actual.fields), len(expected.fields))
	}
	for i, expected := range expected.fields {
		actual := actual.fields[i]
		if actual.key != expected.key {
			return reportUnexpected("key", actual.key, expected.key)
		}

		switch expected := expected.val.(type) {
		case object:
			if err := assertObject(actual.val.(object), expected); err != nil {
				return reportUnexpected("val", actual.val, expected)
			}
		default:
			if actual.val != expected {
				return reportUnexpected("val", actual.val, expected)
			}
		}
	}

	return nil
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
