package go_future

import (
	"fmt"
	"sync/atomic"
)

var eventBus = make(chan *func() bool)

type Promise struct {
	id       uint64
	handler  bool
	rev      <-chan interface{}
	errChan  chan error
	doneFlag bool
	done     chan struct{}
}

func (e *eventLoop) newPromise(rev <-chan interface{}, errChan chan error) *Promise {
	currentP := &Promise{id: uint64(atomic.AddInt64(&e.size, 1)), rev: rev, errChan: errChan, done: make(chan struct{})}
	e.sync.Lock()
	e.promiseQueue = append(e.promiseQueue, currentP)
	e.sync.Unlock()
	return currentP
}

func (p *Promise) complete() {
	close(p.done)
}

func (p *Promise) registerHandler() {
	p.handler = true
}

func (p *Promise) Then(fn func(interface{})) *Promise {
	p.registerHandler()
	work := func() bool {
		select {
		default:
			if p.doneFlag {
				return true
			}
			return false
		case val := <-p.rev:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						switch x := r.(type) {
						case error:
							p.errChan <- x
						default:
							p.errChan <- fmt.Errorf("%v", x)
						}
					} else {
						p.doneFlag = true
						p.complete()
					}
				}()
				fn(val)
			}()
			return true
		}
	}
	eventBus <- &work
	return p
}

func (p *Promise) Catch(fn func(err error)) {
	p.registerHandler()
	work := func() bool {
		select {
		default:
			if p.doneFlag {
				return true
			}
			return false
		case err := <-p.errChan:
			p.doneFlag = true
			go fn(err)
			p.complete()
			return true
		}
	}
	eventBus <- &work
}

func (p *Promise) Await() (any, error) {
	defer p.complete()
	p.registerHandler()
	select {

	case err := <-p.errChan:
		return nil, err
	case rev := <-p.rev:
		return rev, nil
	}
}
