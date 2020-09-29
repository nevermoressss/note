#### 网络轮询器
```text
为了提高 I/O 多路复用的性能，不同的操作系统也都实现了自己的 I/O 多路复用函数，
例如：epoll、kqueue 和 evport 等。
Go 语言为了提高在不同操作系统上的 I/O 操作性能，使用平台特定的函数实现了多个版本的网络轮询模块
src/runtime/netpoll_epoll.go
src/runtime/netpoll_kqueue.go
src/runtime/netpoll_solaris.go
src/runtime/netpoll_windows.go
src/runtime/netpoll_aix.go
src/runtime/netpoll_fake.go
```
在runtime/netpoll.go 文件中描述了需要实现的方法
```go
func netpollinit()
    //初始化网络轮询器
func netpollopen(fd uintptr, pd *pollDesc) int32
   //监听文件描述符上的边缘触发事件，创建事件并加入监听
func netpoll(delta int64) gList
    //如果参数小于 0，无限期等待文件描述符就绪；
    //如果参数等于 0，非阻塞地轮询网络；
    //如果参数大于 0，阻塞特定时间轮询网络；
func netpollBreak()
    //唤醒网络轮询器
func netpollIsPollDescriptor(fd uintptr) bool
    //判断文件描述符是否被轮询器使用
```
对应文件描述符的结构体
```go
//go:notinheap
type pollDesc struct {
	link *pollDesc // 链表
	lock    mutex  // 锁
	fd      uintptr // 地址
	closing bool
	everr   bool    // marks event scanning error happened
	user    uint32  // user settable cookie
	rseq    uintptr // 防止过期的读取
	rg      uintptr // pdReady、pdWait、等待文件描述符可读或者可写的 Goroutine 以及 nil
	rt      timer   // 用于等待文件描述符的计时器；
	rd      int64   // 等待文件描述符可读或者可写的截止日期
	wseq    uintptr // 防止过期的写
	wg      uintptr // pdReady、pdWait、等待文件描述符可读或者可写的 Goroutine 以及 nil
	wt      timer   // 用于等待文件描述符的计时器；
	wd      int64   // 等待文件描述符可读或者可写的截止日期
}
```
当我们在文件描述符上执行读写操作时，如果文件描述符不可读或者不可写，当前 Goroutine 就会执行 runtime.poll_runtime_pollWait 检查 runtime.pollDesc 的状态并调用 runtime.netpollblock 等待文件描述符的可读或者可写
```go
func poll_runtime_pollWait(pd *pollDesc, mode int) int {
	// 检查状态
	errcode := netpollcheckerr(pd, int32(mode))
	if errcode != pollNoError {
		return errcode
	}
	// As for now only Solaris, illumos, and AIX use level-triggered IO.
	if GOOS == "solaris" || GOOS == "illumos" || GOOS == "aix" {
		netpollarm(pd, mode)
	}
	// runtime.netpollblock 是 Goroutine 等待 I/O 事件的关键函数，
	// 它会使用运行时提供的 runtime.gopark 让出当前线程，将 Goroutine 转换到休眠状态并等待运行时的唤醒。
	for !netpollblock(pd, int32(mode), false) {
		errcode = netpollcheckerr(pd, int32(mode))
		if errcode != pollNoError {
			return errcode
		}
	}
	return pollNoError
}
```
轮询等待
```text
Go 语言的运行时会在调度或者系统监控中调用 runtime.netpoll 轮询网络，该函数的执行过程可以分成以下几个部分：

1、根据传入的 delay 计算 epoll 系统调用需要等待的时间；
2、调用 epollwait 等待可读或者可写事件的发生；
3、在循环中依次处理 epollevent 事件；
```
小结
```text
网络轮询器并不是由运行时中的某一个线程独立运行的，运行时中的调度和系统调用会通过 runtime.netpoll 与网络轮询器交换消息，获取待执行的 Goroutine 列表，并将待执行的 Goroutine 加入运行队列等待处理。

所有的文件 I/O、网络 I/O 和计时器都是由网络轮询器管理的
```