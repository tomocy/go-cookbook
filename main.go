package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

func main() {
	ps := newpubsub()
	defer ps.close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ps.serve(ctx)

	subs := make([]chan string, 3)
	for i := range subs {
		subs[i] = make(chan string)
	}

	for _, sub := range subs {
		ps.subscribe(sub)
	}

	var wg sync.WaitGroup
	for i, sub := range subs {
		wg.Add(1)
		go func(msgCh <-chan string, cnt int) {
			defer wg.Done()
			for msg := range msgCh {
				fmt.Println(strings.Repeat(msg, cnt))
			}
		}(sub, i+1)
	}

	for i := 0; i < 10; i++ {
		ps.publish(fmt.Sprint(i))
	}

	for _, sub := range subs {
		ps.unsubscribe(sub)
		close(sub)
	}

	wg.Wait()
}

func newpubsub() *pubsub {
	return &pubsub{
		subs:    make(map[sub]struct{}),
		subCh:   make(chan sub),
		unsubCh: make(chan sub),
		msgCh:   make(chan string),
	}
}

type pubsub struct {
	subs    map[sub]struct{}
	subCh   chan sub
	unsubCh chan sub
	msgCh   chan string
}

func (ps *pubsub) subscribe(sub sub) {
	ps.subCh <- sub
}

func (ps *pubsub) unsubscribe(sub sub) {
	ps.unsubCh <- sub
}

func (ps *pubsub) publish(msg string) {
	ps.msgCh <- msg
}

func (ps *pubsub) serve(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ch := <-ps.subCh:
			ps.subs[ch] = struct{}{}
		case ch := <-ps.unsubCh:
			delete(ps.subs, ch)
		case msg := <-ps.msgCh:
			for sub := range ps.subs {
				sub <- msg
			}
		}
	}
}

func (ps *pubsub) close() {
	close(ps.subCh)
	close(ps.unsubCh)
	close(ps.msgCh)
}

type sub chan<- string
