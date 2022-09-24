package main

import (
	"fmt"
	"github.com/alob-mtc/go-promise/eventloop"
	"time"
)

var GlobalEventLoop = eventloop.New()

func main() {
	GlobalEventLoop.Main(func() {
		result := GetUserName(2)

		result.Then(func(x interface{}) {
			fmt.Println("2 : user:", x)
		})

		fmt.Println("run before promise returns")

		GetUserName(0).Then(func(x interface{}) {
			fmt.Println("0 : user:", x)
		}).Catch(func(err error) {
			fmt.Println("0 : err:", err)
		})

		GetUserName(10).Then(func(x interface{}) {
			fmt.Println("10 : user:", x)
		}).Catch(func(err error) {
			fmt.Println("10 : err:", err)
		})

		//	await
		syncResult1 := GlobalEventLoop.Await(GetUserName(4))
		fmt.Println("4 : user:", syncResult1)

		syncResult2 := GlobalEventLoop.Await(GetUserName(1))
		fmt.Println("1 : user:", syncResult2)

		fmt.Println("done")

	})
}

func GetUserName(id time.Duration) *eventloop.Promise {
	result := make(chan interface{})
	errChan := make(chan error)

	go func() {
		<-time.After(time.Second * id)
		if id == 0 {
			errChan <- fmt.Errorf("some error id(%s)", id)
		} else {
			result <- fmt.Sprintf("id(%s): Test User", id)
		}
	}()

	return GlobalEventLoop.NewPromise(result, errChan)
}
