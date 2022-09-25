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

		//	await
		syncResult1 := GlobalEventLoop.Await(GetUserName(4))
		fmt.Println("4 : user:", syncResult1)

		syncResult2 := GlobalEventLoop.Await(GetUserName(1))
		fmt.Println("1 : user:", syncResult2)

		asyncResult := GetUserName(6)
		GetUserName(3)

		fmt.Println("asyncResult", GlobalEventLoop.Await(asyncResult))

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
