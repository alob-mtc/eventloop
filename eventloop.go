package go_future

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	once            sync.Once
	GlobalEventLoop *eventLoop
)

func Init() {
	queue := []func() bool{}
	var lock sync.Mutex
	once.Do(func() {
		GlobalEventLoop = &eventLoop{promiseQueue: make([]*Promise, 0)}
	})
	// TODO: add evnt loop
	go func() {
		for work := range eventBus {
			lock.Lock()
			queue = append(queue, *work)
			lock.Unlock()
		}
	}()

	go func() {
		for {
			n := len(queue)
			retry := []func() bool{}
			for i := 0; i < n; i++ {
				currentWork := queue[i]

				if done := currentWork(); !done {
					retry = append(retry, currentWork)
				}
			}
			lock.Lock()
			queue = queue[n:]
			queue = append(queue, retry...)
			lock.Unlock()
		}
	}()
}

func GetGlobalEventLoop() Future {
	return GlobalEventLoop
}

type Future interface {
	Await(currentP *Promise) (interface{}, error)
	Async(fn func() (interface{}, error)) *Promise
	Run()
}

type eventLoop struct {
	promiseQueue []*Promise
	size         int64
	sync         sync.Mutex
}

func (e *eventLoop) Await(currentP *Promise) (interface{}, error) {
	defer currentP.Done()
	currentP.RegisterHandler()
	select {

	case err := <-currentP.errChan:
		return nil, err
	case rev := <-currentP.rev:
		return rev, nil
	}
}

func (e *eventLoop) Async(fn func() (interface{}, error)) *Promise {
	resultChan := make(chan interface{})
	errChan := make(chan error)
	p := e.newPromise(resultChan, errChan)
	go func() {
		recoveryHandler := promiseRecovery(resultChan, errChan)
		defer func() {
			if r := recover(); r != nil {
				switch x := r.(type) {
				case error:
					recoveryHandler(nil, x)
				default:
					recoveryHandler(nil, fmt.Errorf("%v", x))
				}
			}
		}()
		result, err := fn()
		recoveryHandler(result, err)
	}()
	return p
}

func promiseRecovery(resultChan chan interface{}, errChan chan error) func(result interface{}, err error) {
	return func(result interface{}, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()
		if err != nil {
			select {
			case errChan <- err:
			case <-ctx.Done():
			}
			return
		}

		select {
		case resultChan <- result:
		case <-ctx.Done():
		}
	}
}

func (e *eventLoop) Run() {
	//await all promises
	e.awaitAll()
}

func (e *eventLoop) awaitAll() {
	init := true
outer:
	for {
		n := len(e.promiseQueue)
		if init && n == 0 {
			return
		}
		for i := n - 1; i >= 0; i-- {
			e.sync.Lock()
			p := e.promiseQueue[i]
			e.sync.Unlock()
			if p.handler {
				<-p.done
			}
			if currentN := int(atomic.LoadInt64(&e.size)); i == 0 && !(currentN > n) {
				close(eventBus)
				break outer
			}
		}
		// clean up memory (promise)
		e.sync.Lock()
		e.promiseQueue = e.promiseQueue[n:]
		atomic.AddInt64(&e.size, int64(-n))
		e.sync.Unlock()
		init = false
	}
}
