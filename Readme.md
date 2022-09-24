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
```
#### result

```shell
run before promise returns
0 : err: some error id(0s)
2 : user: id(2ns): Test User
4 : user: id(4ns): Test User
1 : user: id(1ns): Test User
done
10 : user: id(10ns): Test User

```


> just a fun project, we might just learn something