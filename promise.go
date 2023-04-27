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
	err     bool
	done    chan struct{}
}

func (e *eventLoop) newPromise(rev <-chan interface{}, errChan chan error) *Promise {
	currentP := &Promise{id: uint64(atomic.AddInt64(&e.size, 1)), rev: rev, errChan: errChan, done: make(chan struct{})}
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
		for {
			select {
			default:
				if p.err {
					return
				}
			case val := <-p.rev:
				func() {
					defer func() {
						if r := recover(); r != nil {
							switch x := r.(type) {
							case error:
								p.errChan <- x
							default:
								p.errChan <- fmt.Errorf("%v", x)
							}
						} else {
							p.err = true
							p.Done()
						}
					}()
					fn(val)
				}()
				return
			}
		}
	}()
	return p
}

func (p *Promise) Catch(fn func(err error)) {
	p.RegisterHandler()
	go func() {
		for {
			select {
			default:
				if p.err {
					return
				}
			case err := <-p.errChan:
				fn(err)
				p.err = true
				p.Done()
			}
		}
	}()
}
