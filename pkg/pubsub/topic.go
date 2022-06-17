package pubsub

import (
	"context"
	"sync"
)

// Topic is an unbounded pub sub. This means that when publishing an event,
// all subscribers will receive the event but a go routine will be started for every subscriber to consume the event.
// This is useful for situations where you want to send a message to a large number of subscribers but you don't want to
// block the publisher.
// You have to call the close method to release all resources.
type Topic[T any] struct {
	events   []chan<- T
	mu       sync.RWMutex
	cancel   context.CancelFunc
	ctx      context.Context
	wg       sync.WaitGroup
	isClosed bool
}

func NewTopic[T any](parentContext context.Context) *Topic[T] {
	ctx, cancel := context.WithCancel(parentContext)

	return &Topic[T]{
		events: []chan<- T{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (o *Topic[T]) Subscribe() <-chan T {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.isClosed {
		return nil
	}
	if o.events == nil {
		o.events = make([]chan<- T, 0)
	}
	ch := make(chan T)
	o.events = append(o.events, ch)
	return ch
}

func (o *Topic[T]) Publish(evt T) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if o.isClosed || len(o.events) == 0 {
		return
	}

	for _, c := range o.events {
		o.wg.Add(1)
		go func(c chan<- T) {
			defer o.wg.Done()
			select {
			case c <- evt:
			case <-o.ctx.Done():
			}
		}(c)
	}
}

func (o *Topic[T]) Close() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.isClosed {
		return
	}
	o.isClosed = true
	o.cancel()
	o.wg.Wait()
	for _, c := range o.events {
		close(c)
	}
}

func (o *Topic[T]) IsClosed() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.isClosed
}
