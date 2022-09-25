package eventloop

import (
	"errors"
)

type EventLoop struct {
	promiseQueue []*Promise
}

func New() *EventLoop {
	return &EventLoop{promiseQueue: []*Promise{}}
}

func (e *EventLoop) Await(currentP *Promise) interface{} {
	defer currentP.Done()
	currentP.RegisterHandler()
	val := <-currentP.rev
	return val
}

func (e *EventLoop) Async(fn func() (interface{}, error)) *Promise {
	resultChan := make(chan interface{})
	errChan := make(chan error)
	p := e.NewPromise(resultChan, errChan)
	go func() {
		result, err := fn()
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()
	return p
}

func (e *EventLoop) Main(fn func()) {
	fn()
	//await all promises
	for i := len(e.promiseQueue) - 1; i >= 0; i-- {
		p := e.promiseQueue[i]
		if p.handler {
			<-p.done
		}
	}
}

//Promise

type Promise struct {
	handler bool
	rev     <-chan interface{}
	errChan chan error
	err     chan struct{}
	done    chan struct{}
}

func (e *EventLoop) NewPromise(rev <-chan interface{}, errChan chan error) *Promise {
	currentP := &Promise{rev: rev, errChan: errChan, done: make(chan struct{}), err: make(chan struct{})}
	e.promiseQueue = append(e.promiseQueue, currentP)
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
					case string:
						p.errChan <- errors.New(x)
					case error:
						p.errChan <- x
					default:
						p.errChan <- errors.New("unknown error")
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
