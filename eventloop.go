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
	once.Do(func() {
		GlobalEventLoop = &eventLoop{promiseQueue: []*Promise{}, keepAlive: true}
	})
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
	size         uint64
	keepAlive    bool
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
	for {
		n := len(e.promiseQueue)
		for i := n - 1; i >= 0; i-- {
			e.sync.Lock()
			p := e.promiseQueue[i]
			e.sync.Unlock()
			if p.handler {
				<-p.done
			}
			if currentN := int(atomic.LoadUint64(&e.size)); i == 0 && !(currentN > n) {
				break
			}
		}
		// clean up memory (promise)
		e.sync.Lock()
		e.promiseQueue = e.promiseQueue[n:]
		atomic.StoreUint64(&e.size, e.size-uint64(n))
		e.sync.Unlock()
	}
}
