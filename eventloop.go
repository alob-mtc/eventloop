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

func InitGlobalEventLoop() Future {
	once.Do(func() {
		var queue []func() bool
		var lock sync.Mutex
		// event loop
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
				var retry []func() bool
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
		GlobalEventLoop = &eventLoop{promiseQueue: make([]*Promise, 0)}
	})
	return GlobalEventLoop
}

type Future interface {
	Async(fn func() (any, error)) *Promise
	Run()
}

type eventLoop struct {
	promiseQueue []*Promise
	size         int64
	sync         sync.Mutex
}

func (e *eventLoop) Run() {
	//await all promises
	e.awaitAll()
}

func (e *eventLoop) Async(fn func() (any, error)) *Promise {
	resultChan := make(chan any)
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

func promiseRecovery(resultChan chan any, errChan chan error) func(result any, err error) {
	return func(result any, err error) {
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
