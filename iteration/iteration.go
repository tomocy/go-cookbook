package iteration

import (
	"fmt"
	"io"
)

type withFor struct{}

func (p withFor) print(w io.Writer, max int) {
	for i := 2; i <= max; i += 2 {
		fmt.Fprintln(w, i)
	}
}

type withCallback struct{}

func (p withCallback) print(w io.Writer, max int) {
	p.rangeEvenNumber(max, func(n int) {
		fmt.Fprintln(w, n)
	})
}

func (p withCallback) rangeEvenNumber(max int, fn func(int)) {
	for i := 2; i <= max; i += 2 {
		fn(i)
	}
}

type withIter struct{}

func (p withIter) print(w io.Writer, max int) {
	i := iter{
		max: max,
	}
	for i.next() {
		if i.val%2 != 0 {
			continue
		}

		fmt.Fprintln(w, i.val)
	}
}

type iter struct {
	max int
	val int
	err error
}

func (i *iter) next() bool {
	if i.err != nil {
		return false
	}

	i.val++

	return i.val <= i.max
}

type withChan struct{}

func (p withChan) print(w io.Writer, max int) {
	nums := p.evenNums(max)
	for n := range nums {
		fmt.Fprintln(w, n)
	}
}

func (p withChan) evenNums(max int) <-chan int {
	nums := make(chan int)
	go func() {
		defer close(nums)

		for i := 2; i <= max; i += 2 {
			nums <- i
		}
	}()

	return nums
}

type evenNumberPrinter interface {
	print(io.Writer, int)
}
