package eventloop

type EventLoop struct {
	promiseQueue []*Promise
}

func (e *EventLoop) Await(currentP *Promise) interface{} {
	val := <-currentP.rev
	go func() { currentP.done <- struct{}{} }()
	return val
}

func (e *EventLoop) Main(fn func()) {
	fn()
	//TODO: await all promises
	for i := len(e.promiseQueue) - 1; i >= 0; i-- {
		p := e.promiseQueue[i]
		<-p.done
	}
}

func New() *EventLoop {
	return &EventLoop{promiseQueue: []*Promise{}}
}

type Promise struct {
	rev     <-chan interface{}
	errChan <-chan error
	err     chan struct{}
	done    chan struct{}
}

func (e *EventLoop) NewPromise(rev <-chan interface{}, errChan <-chan error) *Promise {
	currentP := &Promise{rev: rev, errChan: errChan, done: make(chan struct{}), err: make(chan struct{})}
	e.promiseQueue = append(e.promiseQueue, currentP)
	return currentP
}

func (p *Promise) Then(fn func(interface{})) *Promise {
	go func() {
		select {
		case <-p.err:
		case val := <-p.rev:
			close(p.err)
			fn(val)
			p.done <- struct{}{}
		}
	}()
	return p
}

func (p *Promise) Catch(fn func(err error)) {
	go func() {
		select {
		case <-p.err:
		case err := <-p.errChan:
			close(p.err)
			fn(err)
			p.done <- struct{}{}
		}
	}()
}
