package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	d := newDeamon()
	go d.serve()

	if err := d.sendRequest(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %s\n", err)
		os.Exit(1)
	}
}

func newDeamon() *deamon {
	return &deamon{
		reqCh: make(chan respCh),
	}
}

type deamon struct {
	reqCh chan respCh
}

func (d *deamon) serve() {
	for {
		select {
		case respCh := <-d.reqCh:
			go func() {
				respCh <- d.handle()
			}()
		}
	}
}

func (d *deamon) handle() error {
	var err error
	rand.Seed(time.Now().UnixNano())
	if rand.Float32() < 0.5 {
		err = errors.New("failed to handle")
	}

	return err
}

func (d *deamon) sendRequest() error {
	ch := make(chan error)
	defer close(ch)

	d.reqCh <- ch

	return <-ch
}

type respCh chan error
