#### defer
```text
Go 语言的 defer 会在当前函数或者方法返回之前执行传入的函数。
它会经常被用于关闭文件描述符、关闭数据库连接以及解锁资源。
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
deferproc:runtime.deferproc
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