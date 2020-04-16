package batch

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	type r struct {
		begin, end int
	}

	length, size := 100, 10

	expected := make([]r, size)
	for i := range expected {
		begin := size * i
		expected[i] = r{
			begin: begin,
			end:   begin + size - 1,
		}
	}

	var actual []r
	if err := batch(length, size, func(begin, end int) error {
		actual = append(actual, r{
			begin: begin,
			end:   end,
		})

		return nil
	}); err != nil {
		t.Errorf("should have batched: %s", err)
		return
	}

	if len(actual) != len(expected) {
		t.Errorf("should have batch the expected times: %s", reportUnexpected("len of ranges", len(actual), len(expected)))
		return
	}
	for i := range expected {
		if actual[i].begin != expected[i].begin {
			t.Errorf("should have batch at %d expectedly: %s", i, reportUnexpected("begin", actual[i].begin, expected[i].begin))
			return
		}
		if actual[i].end != expected[i].end {
			t.Errorf("should have batch at %d expectedly: %s", i, reportUnexpected("end", actual[i].end, expected[i].end))
			return
		}
	}
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
