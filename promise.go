package go_future

import (
	"fmt"
	"sync/atomic"
)

type Promise struct {
	id      uint64
	handler bool
	rev     <-chan interface{}
	errChan chan error
	err     chan struct{}
	done    chan struct{}
}

func (e *eventLoop) newPromise(rev <-chan interface{}, errChan chan error) *Promise {
	if atomic.LoadUint64(&e.size) == 0 {
		defer close(e.signal)
	}
	currentP := &Promise{id: atomic.AddUint64(&e.size, 1), rev: rev, errChan: errChan, done: make(chan struct{}), err: make(chan struct{})}
	e.sync.Lock()
	e.promiseQueue = append(e.promiseQueue, currentP)
	e.sync.Unlock()
	return currentP
}

func (p *Promise) Done() {
	close(p.done)
}

func (p *Promise) RegisterHandler() {
	p.handler = true
}

func (p *Promise) Then(fn func(interface{})) *Promise {
	p.RegisterHandler()
	go func() {
		select {
		case <-p.err:
		case val := <-p.rev:
			defer func() {
				if r := recover(); r != nil {
					switch x := r.(type) {
					case error:
						p.errChan <- x
					default:
						p.errChan <- fmt.Errorf("%v", x)
					}
				} else {
					close(p.err)
					p.Done()
				}
			}()
			fn(val)
		}
	}()
	return p
}

func (p *Promise) Catch(fn func(err error)) {
	p.RegisterHandler()
	go func() {
		select {
		case <-p.err:
		case err := <-p.errChan:
			close(p.err)
			fn(err)
			p.Done()
		}
	}()
}
