package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	for line := range transform(fetch()) {
		fmt.Println(line)
	}
}

func transform(srcCh <-chan string) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		for src := range srcCh {
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			ch <- fmt.Sprintf("%s was fetched.", src)
		}
	}()

	return ch
}

func fetch() <-chan string {
	chCh := make(chan chan string, 10)

	go func() {
		defer close(chCh)

		for i := 0; i < 1000; i++ {
			ch := make(chan string)
			chCh <- ch

			go func(i int) {
				defer close(ch)

				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				ch <- fmt.Sprintf("line:%d", i)
			}(i)
		}
	}()

	ch := make(chan string)

	go func() {
		defer close(ch)

		for srcCh := range chCh {
			ch <- <-srcCh
		}
	}()

	return ch
}
