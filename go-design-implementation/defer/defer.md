#### defer
```text
Go 语言的 defer 会在当前函数或者方法返回之前执行传入的函数。
它会经常被用于关闭文件描述符、关闭数据库连接以及解锁资源。
```
Go 语言 defer 语句的三种机制（来源网络）
```text
Golang 的 1.13 版本 与 1.14 版本对 defer 进行了两次优化
堆上分配
    在 Golang 1.13 之前的版本中，所有 defer 都是在堆上分配，该机制在编译时会进行两个步骤：
        1、在 defer 语句的位置插入 runtime.deferproc，当被执行时，延迟调用会被保存为一个 _defer 记录，并将被延迟调用的入口地址及其参数复制保存，存入 Goroutine 的调用链表中。
        2、在函数返回之前的位置插入 runtime.deferreturn，当被执行时，会将延迟调用从 Goroutine 链表中取出并执行，多个延迟调用则以 jmpdefer 尾递归调用方式连续执行。
    这种机制的主要性能问题存在于每个 defer 语句产生记录时的内存分配，以及记录参数和完成调用时参数移动的系统调用开销。
栈上分配
    Go 1.13 版本新加入 deferprocStack 实现了在栈上分配的形式来取代 deferproc，相比后者，栈上分配在函数返回后 _defer 便得到释放，省去了内存分配时产生的性能开销，只需适当维护 _defer 的链表即可。
    
    编译器有自己的逻辑去选择使用 deferproc 还是 deferprocStack，大部分情况下都会使用后者，性能会提升约 30%。不过在 defer 语句出现在了循环语句里，或者无法执行更高阶的编译器优化时，亦或者同一个函数中使用了过多的 defer 时，依然会使用 deferproc。
开放编码
    Go 1.14 版本继续加入了开发编码（open coded），该机制会将延迟调用直接插入函数返回之前，省去了运行时的 deferproc 或 deferprocStack 操作，在运行时的 deferreturn 也不会进行尾递归调用，而是直接在一个循环中遍历所有延迟函数执行。
    
    这种机制使得 defer 的开销几乎可以忽略，唯一的运行时成本就是存储参与延迟调用的相关信息，不过使用此机制需要一些条件：
    
        1、没有禁用编译器优化，即没有设置 -gcflags "-N"；
        2、函数内 defer 的数量不超过 8 个，且返回语句与延迟语句个数的乘积不超过 15；
        3、defer 不是在循环语句中。
        
    该机制还引入了一种元素 —— 延迟比特（defer bit），用于运行时记录每个 defer 是否被执行（尤其是在条件判断分支中的 defer），从而便于判断最后的延迟调用该执行哪些函数。
    延迟比特的原理：
    同一个函数内每出现一个 defer 都会为其分配 1 个比特，如果被执行到则设为 1，否则设为 0，当到达函数返回之前需要判断延迟调用时，则用掩码判断每个位置的比特，若为 1 则调用延迟函数，否则跳过。

```
_defer:runtime/defer
```go
type _defer struct {
	siz     int32 // 包含入参和出参的大小
	started bool
	heap    bool
	openDefer bool
	sp        uintptr  // 栈指针
	pc        uintptr  // 程序计数器
	fn        *funcval // 延迟处理的方法
	_panic    *_panic  // 是否触发过panic
	link      *_defer  // 下一个defer
	// If openDefer is true, the fields below record values about the stack
	// frame and associated function that has the open-coded defer(s). sp
	// above will be the sp for the frame, and pc will be address of the
	// deferreturn call in the function.
	fd   unsafe.Pointer // funcdata for the function associated with the frame
	varp uintptr        // value of varp for the stack frame
	// framepc is the current pc associated with the stack frame. Together,
	// with sp above (which is the sp associated with the stack frame),
	// framepc/sp can be used as pc/sp pair to continue a stack trace via
	// gentraceback().
	framepc uintptr
}
```
编译器对defer的调整
```text
将 defer 关键字都转换成 runtime.deferproc 函数
为所有调用 defer 的函数末尾插入 runtime.deferreturn 的函数调用
```
deferproc:runtime.deferproc  (在堆上分配)    
```go
func deferproc(siz int32, fn *funcval) {
	// 判断当前的g是不是实际运行的g
	gp := getg()
	if gp.m.curg != gp {
		// go code on the system stack can't defer
		throw("defer on system stack")
	}
	// 获取栈指针
	sp := getcallersp()
	// 
	argp := uintptr(unsafe.Pointer(&fn)) + unsafe.Sizeof(fn)
	// pc 计数器
	callerpc := getcallerpc()
    // 新建defer结构体
	d := newdefer(siz)
	if d._panic != nil {
		throw("deferproc: d.panic != nil after newdefer")
	}
	// 在队头加入
	d.link = gp._defer
	gp._defer = d
	// 设置方法和sp pc
	d.fn = fn
	d.pc = callerpc
	d.sp = sp
	// 参数放的位置
	switch siz {
	case 0:
		// Do nothing.
	case sys.PtrSize:
		*(*uintptr)(deferArgs(d)) = *(*uintptr)(unsafe.Pointer(argp))
	default:
		memmove(deferArgs(d), unsafe.Pointer(argp), uintptr(siz))
	}

	//函数的作用是避免无限递归调用 runtime.deferreturn，它是唯一一个不会触发由延迟调用的函数了。
	return0()
}
```
deferreturn:runtime.deferreturn
```go
//go:nosplit
func deferreturn(arg0 uintptr) {
	// 获取g
	gp := getg()
	// 没有defer方法 return
	d := gp._defer
	if d == nil {
		return
	}
	// 获取栈指针
	sp := getcallersp()
	if d.sp != sp {
		return
	}
	// 有defer 是否开放编码
	if d.openDefer {
		done := runOpenDeferFrame(gp, d)
		if !done {
			throw("unfinished open-coded defers in deferreturn")
		}
		gp._defer = d.link
		freedefer(d)
		return
	}
	// 获取参数
	switch d.siz {
	case 0:
		// Do nothing.
	case sys.PtrSize:
		*(*uintptr)(unsafe.Pointer(&arg0)) = *(*uintptr)(deferArgs(d))
	default:
		memmove(unsafe.Pointer(&arg0), deferArgs(d), uintptr(d.siz))
	}
	fn := d.fn
	d.fn = nil
	gp._defer = d.link
	freedefer(d)
	_ = fn.fn
	// 执行方法
	jmpdefer(fn, uintptr(unsafe.Pointer(&arg0)))
}
```