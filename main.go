package main

import (
	"fmt"
	"time"

	"github.com/alob-mtc/go-promise/eventloop"
)

var GlobalEventLoop *eventloop.EventLoop

func init() {
	eventloop.Init()
	GlobalEventLoop = eventloop.GetGlobalEventLoop()
}

func GetUserName(id time.Duration) *eventloop.Promise {
	return GlobalEventLoop.Async(func() (interface{}, error) {
		<-time.After(time.Second * id)
		if id == 0 {
			return nil, fmt.Errorf("some error id(%s)", id)
		}
		return fmt.Sprintf("id(%s): Test User", id), nil
	})
}

func GetUserNameWithPanic() *eventloop.Promise {
	return GlobalEventLoop.Async(func() (interface{}, error) {
		<-time.After(time.Second * 2)
		panic("panic attack")
	})
}

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

		GetUserName(5).Then(func(x interface{}) {
			fmt.Println("5 : user:", x)
			panic("a panic attack")
		}).Catch(func(err error) {
			fmt.Println("5 : err:", err)
		})

		GetUserName(15).Then(func(x interface{}) {
			fmt.Println("15 : user:", x)
		}).Catch(func(err error) {
			fmt.Println("15 : err:", err)
		})

		//promise with panic
		GetUserNameWithPanic().Then(func(i interface{}) {
			fmt.Println("Then block never gets triggered")
		}).Catch(func(err error) {
			fmt.Println("GetUserNameWithPanic err: ", err)
		})

		//	await
		syncResult1, err := GlobalEventLoop.Await(GetUserName(4))
		fmt.Printf("syncResult1 - value: %v, err: %v\n", syncResult1, err)

		syncResult2, err := GlobalEventLoop.Await(GetUserName(1))
		fmt.Printf("syncResult2 - value: %v, err: %v\n", syncResult2, err)

		asyncResult := GetUserName(0)
		GetUserName(3)

		syncResult, err := GlobalEventLoop.Await(asyncResult)
		fmt.Printf("asyncResult - value: %v, err: %v\n", syncResult, err)

		fmt.Println("done")

		//nested promise
		GlobalEventLoop.Async(func() (interface{}, error) {
			fmt.Println("outer async")
			GlobalEventLoop.Async(func() (interface{}, error) {
				fmt.Println("inner async")
				return nil, nil
			}).Then(func(_ interface{}) {
				fmt.Println("resolved inner promise")
			})
			<-time.After(time.Second * 2)
			return nil, nil
		}).Then(func(_ interface{}) {
			fmt.Println("resolved outer promise")
		})
	})
}
