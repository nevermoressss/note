##### 单线程调度器 · 0.x
只包含 40 多行代码；

程序中只能存在一个活跃线程，由 G-M 模型组成；

##### 多线程调度器 · 1.0
允许运行多线程的程序；

全局锁导致竞争严重；

##### 任务窃取调度器 · 1.1
引入了处理器 P，构成了目前的 G-M-P 模型；

在处理器 P 的基础上实现了基于工作窃取的调度器；

在某些情况下，Goroutine 不会让出线程，进而造成饥饿问题；

时间过长的垃圾回收（Stop-the-world，STW）会导致程序长时间无法工作；

##### 抢占式调度器 · 1.2 ~ 至今
###### 基于协作的抢占式调度器 - 1.2 ~ 1.13

通过编译器在函数调用时插入抢占检查指令，在函数调用时检查当前 Goroutine 是否发起了抢占请求，实现基于协作的抢占式调度
；
Goroutine 可能会因为垃圾回收和循环长时间占用资源导致程序暂停；

###### 基于信号的抢占式调度器 - 1.14 ~ 至今
实现基于信号的真抢占式调度；

垃圾回收在扫描栈时会触发抢占调度；

抢占的时间点不够多，还不能覆盖全部的边缘情况；

###### 非均匀存储访问调度器 · 提案
对运行时的各种资源进行分区；

实现非常复杂，到今天还没有提上日程；


##### 协作的抢占式调度器

1 编译器会在调用函数前插入 runtime.morestack；

2 Go 语言运行时会在垃圾回收暂停程序、系统监控发现 Goroutine 运行超过 10ms 时发出抢占请求 StackPreempt；

3 当发生函数调用时，可能会执行编译器插入的 runtime.morestack 函数，它调用的 runtime.newstack 会检查 Goroutine 的 stackguard0 字段是否为 StackPreempt；

4 如果 stackguard0 是 StackPreempt，就会触发抢占让出当前线程；
##### 非协作的抢占式调度
1 程序启动时，在 runtime.sighandler 函数中注册 SIGURG 信号的处理函数 runtime.doSigPreempt；

2 在触发垃圾回收的栈扫描时会调用 runtime.suspendG 挂起 Goroutine，该函数会执行下面的逻辑：

    1.将 _Grunning 状态的 Goroutine 标记成可以被抢占，即将 preemptStop 设置成 true；
    
    2.调用 runtime.preemptM 触发抢占
3 runtime.preemptM 会调用 runtime.signalM 向线程发送信号 SIGURG；

4 操作系统会中断正在运行的线程并执行预先注册的信号处理函数 runtime.doSigPreempt；

5 runtime.doSigPreempt 函数会处理抢占信号，获取当前的 SP 和 PC 寄存器并调用 runtime.sigctxt.pushCall

6 runtime.sigctxt.pushCall 会修改寄存器并在程序回到用户态时执行runtime.asyncPreempt；

7 汇编指令 runtime.asyncPreempt 会调用运行时函数 runtime.asyncPreempt2；

8 runtime.asyncPreempt2 会调用 runtime.preemptPark；

9 runtime.preemptPark 会修改当前 Goroutine 的状态到 _Gpreempted 并调用 runtime.schedule 让当前函数陷入休眠并让出线程，调度器会选择其它的 Goroutine 继续执行；
    
```go
// src/runtime/runtime2.go
type m struct {
	g0          *g			// 用于执行调度指令的 goroutine
	gsignal     *g			// 处理 signal 的 g
	tls         [6]uintptr	// 线程本地存储
	curg        *g			// 当前运行的用户 goroutine
	p           puintptr	// 执行 go 代码时持有的 p (如果没有执行则为 nil)
	spinning    bool		// m 当前没有运行 work 且正处于寻找 work 的活跃状态
	cgoCallers  *cgoCallers	// cgo 调用崩溃的 cgo 回溯
	alllink     *m			// 在 allm 上
	mcache      *mcache

	...
}

type p struct {
	id           int32
	status       uint32 // p 的状态 pidle/prunning/...
	link         puintptr
	m            muintptr   // 反向链接到关联的 m （nil 则表示 idle）
	mcache       *mcache
	pcache       pageCache
	deferpool    [5][]*_defer // 不同大小的可用的 defer 结构池
	deferpoolbuf [5][32]*_defer
	runqhead     uint32	// 可运行的 goroutine 队列，可无锁访问
	runqtail     uint32
	runq         [256]guintptr
	runnext      guintptr
	timersLock   mutex
	timers       []*timer
	preempt      bool
	...
}

type g struct {
	stack struct {
		lo uintptr
		hi uintptr
	} 							// 栈内存：[stack.lo, stack.hi)
	stackguard0	uintptr
	stackguard1 uintptr

	_panic       *_panic
	_defer       *_defer
	m            *m				// 当前的 m
	sched        gobuf
	stktopsp     uintptr		// 期望 sp 位于栈顶，用于回溯检查
	param        unsafe.Pointer // wakeup 唤醒时候传递的参数
	atomicstatus uint32
	goid         int64
	preempt      bool       	// 抢占信号，stackguard0 = stackpreempt 的副本
	timer        *timer         // 为 time.Sleep 缓存的计时器

	...
}

type schedt struct {
	lock mutex

	pidle      puintptr	// 空闲 p 链表
	npidle     uint32	// 空闲 p 数量
	nmspinning uint32	// 自旋状态的 M 的数量
	runq       gQueue	// 全局 runnable G 队列
	runqsize   int32
	gFree struct {		// 有效 dead G 的全局缓存.
		lock    mutex
		stack   gList	// 包含栈的 Gs
		noStack gList	// 没有栈的 Gs
		n       int32
	}
	sudoglock  mutex	// sudog 结构的集中缓存
	sudogcache *sudog
	deferlock  mutex	// 不同大小的有效的 defer 结构的池
	deferpool  [5]*_defer
	
	...
}
```

###### runtime.newproc1 获取 Goroutine 结构体的三种方法

1 当处理器的 Goroutine 列表为空时，会将调度器持有的空闲 Goroutine 转移到当前处理器上，直到 gFree 列表中的 Goroutine 数量达到 32；

2 当处理器的 Goroutine 数量充足时，会从列表头部返回一个新的 Goroutine；

3 当调度器的 gFree 和处理器的 gFree 列表都不存在结构体时，运行时会调用 runtime.malg 初始化一个新的 runtime.g 结构体，如果申请的堆栈大小大于 0，在这里我们会通过 runtime.stackalloc 分配 1KB 的栈空间：

总结：runtime.newproc1 会从处理器或者调度器的缓存中获取新的结构体，也可以调用 runtime.malg 函数创建新的结构体。
###### 运行队列

runtime.runqput 函数会将新创建的 Goroutine 运行队列上，这既可能是全局的运行队列，也可能是处理器本地的运行队列

```go
// runqput尝试将g放置在本地可运行队列中。
// 如果next为false，则runqput将g添加到可运行队列的尾部。
// 如果next为true，则runqput将g放入_p_.runnext中。
// 如果运行队列已满，则runnext将g放入全局队列。
// Executed only by the owner P.
func runqput(_p_ *p, gp *g, next bool) {
	if randomizeScheduler && next && fastrand()%2 == 0 {
		next = false
	}

	if next {
	retryNext:
		oldnext := _p_.runnext
		if !_p_.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {
			goto retryNext
		}
		if oldnext == 0 {
			return
		}
		// Kick the old runnext out to the regular run queue.
		gp = oldnext.ptr()
	}

retry:
	h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
	t := _p_.runqtail
	if t-h < uint32(len(_p_.runq)) {
		_p_.runq[t%uint32(len(_p_.runq))].set(gp)
		atomic.StoreRel(&_p_.runqtail, t+1) // store-release, makes the item available for consumption
		return
	}
	if runqputslow(_p_, gp, h, t) {
		return
	}
	// the queue is not full, now the put above must succeed
	goto retry
}
```

```go
//从本地运行队列、全局运行队列中查找
//从网络轮询器中查找是否有 Goroutine 等待运行
//通过 runtime.runqsteal 函数尝试从其他随机的处理器中窃取待运行的 Goroutine，在该过程中还可能窃取处理器中的计时器；
//总而言之，当前函数一定会返回一个可执行的 Goroutine，如果当前不存在就会阻塞等待。
func schedule() {
 ...
}
```
###### 调度时间点
    1 主动挂起 — runtime.gopark -> runtime.park_m
    2 系统调用 — runtime.exitsyscall -> runtime.exitsyscall0
    3 协作式调度 — runtime.Gosched -> runtime.gosched_m -> runtime.goschedImpl
    4 系统监控 — runtime.sysmon -> runtime.retake -> runtime.preemptone
