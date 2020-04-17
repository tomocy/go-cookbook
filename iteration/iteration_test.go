package iteration

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestPrintEvenNumber(t *testing.T) {
	max := 50

	var w strings.Builder
	for i := 2; i <= max; i += 2 {
		fmt.Fprintln(&w, i)
	}
	expected := w.String()

	tests := map[string]evenNumberPrinter{
		"with for":      withFor{},
		"with callback": withCallback{},
		"with iterator": withIter{},
		"with channel":  withChan{},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			var w bytes.Buffer
			test.print(&w, max)

			if w.String() != expected {
				t.Error(reportUnexpected("even number", w.String(), expected))
				return
			}
		})
	}
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, expected %v", name, actual, expected)
}
