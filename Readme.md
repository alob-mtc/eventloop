Go-Promise

```go

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

		// TODO: nested promises not very functional
		GetUserName(7).Then(func(x interface{}) {
			fmt.Println("7 (1): user:", x)
		}).Then(func(x interface{}) {
			fmt.Println("7 (2) : user:", x)
			panic("another panic attack")
		}).Catch(func(err error) {
			fmt.Println("7 : err:", err)
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

		asyncResult1 := GetUserName(6)
		asyncResult2 := GetUserName(3)

		fmt.Println("asyncResult1", GlobalEventLoop.Await(asyncResult1))
		fmt.Println("asyncResult2", GlobalEventLoop.Await(asyncResult2))

		fmt.Println("done")

	})
}
```

#### result

```shell
run before promise returns
0 : err: some error id(0s)
2 : user: id(2ns): Test User
4 : user: id(4ns): Test User
5 : user: id(5ns): Test User
5 : err: a panic attack
1 : user: id(1ns): Test User
7 (1): user: id(7ns): Test User
asyncResult1 id(6ns): Test User
asyncResult2 id(3ns): Test User
done
15 : user: id(15ns): Test User

```

> just a fun project, we might just learn something
