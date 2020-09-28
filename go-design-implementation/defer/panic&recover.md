#### panic&recover
```text
panic 能够改变程序的控制流，函数调用panic 时会立刻停止执行函数的其他代码，并在执行结束后在当前 Goroutine 中递归执行调用方的延迟函数调用 defer

recover 可以中止 panic 造成的程序崩溃。它是一个只能在 defer 中发挥作用的函数，在其他作用域中调用不会发挥任何作用

panic 只会触发当前 Goroutine 的延迟函数调用；
recover 只有在 defer 函数中调用才会生效；
panic 允许在 defer 中嵌套多次调用；
```
panic 
```go
//go:notinheap
type _panic struct {
	argp      unsafe.Pointer //指向 defer 调用时参数的指针
	arg       interface{}    // 参数
	link      *_panic        // 多个panic用链表连接
	pc        uintptr        // 如果忽略了此紧急情况，将在运行时返回哪里  记录pc
	sp        unsafe.Pointer // 如果忽略了此紧急情况，将在运行时返回哪里  记录sp
	recovered bool           // 是否已经被recover
	aborted   bool           // 是否被强行终止
	goexit    bool           // 是否runtime.goexit
}
```
遇到panic的处理步骤 runtime.gopanic
```text
1 创建新的 runtime._panic 结构并添加到所在 Goroutine _panic 链表的最前面；
2 在循环中不断从当前 Goroutine 的 _defer 中链表获取 runtime._defer 并调用 runtime.reflectcall 运行延迟调用函数；
3 调用 runtime.fatalpanic 中止整个程序；
```
recover：runtime.gorecover
```go
func gorecover(argp uintptr) interface{} {
	gp := getg()
	p := gp._panic
	if p != nil && !p.goexit && !p.recovered && argp == uintptr(p.argp) {
		// 恢复
		p.recovered = true
		// 返回参数
		return p.arg
	}
	return nil
}
```
小结(来源draveness)
```text
1、编译器会负责做转换关键字的工作；
    1.将 panic 和 recover 分别转换成 runtime.gopanic 和 runtime.gorecover；
    2.将 defer 转换成 deferproc 函数；
    3.在调用 defer 的函数末尾调用 deferreturn 函数；
2、在运行过程中遇到 gopanic 方法时，会从 Goroutine 的链表依次取出 _defer 结构体并执行；
3、如果调用延迟执行函数时遇到了 gorecover 就会将 _panic.recovered 标记成 true 并返回 panic 的参数；
    1.在这次调用结束之后，gopanic 会从 _defer 结构体中取出程序计数器 pc 和栈指针 sp 并调用 recovery 函数进行恢复程序；
    2.recovery 会根据传入的 pc 和 sp 跳转回 deferproc；
    3.编译器自动生成的代码会发现 deferproc 的返回值不为 0，这时会跳回 deferreturn 并恢复到正常的执行流程；
4、如果没有遇到 gorecover 就会依次遍历所有的 _defer 结构，并在最后调用 fatalpanic 中止程序、打印 panic 的参数并返回错误码 2；
```