####切片
切片就是动态数组，它的长度并不固定，我们可以随意向切片中追加元素，而切片会在容量不足时自动扩容。
cmd/compile/internal/types.NewSlice
```go
// NewSlice returns the slice Type with element type elem.
func NewSlice(elem *Type) *Type {
	//返回的结构体 TSLICE 中的 Extra 字段是一个只包含切片内元素类型的 Slice{Elem: elem} 结构，
	// 也就是说切片内元素的类型是在编译期间确定的，
	// 编译器确定了类型之后，会将类型存储在 Extra 字段中帮助程序在运行时动态获取。
	if t := elem.Cache.slice; t != nil {
		if t.Elem() != elem {
			Fatalf("elem mismatch")
		}
		return t
	}

	t := New(TSLICE)
	t.Extra = Slice{Elem: elem}
	elem.Cache.slice = t
	return t
}
```
运行时切片由如下的 SliceHeader 结构体表示

reflect.SliceHeader
```go
type SliceHeader struct {
	Data uintptr    //指向数组的指针
	Len  int        //当前切片的长度
	Cap  int        //当前切片的容量
}
```
初始化
```go
arr[0:3] or slice[0:3]      //通过下标的方式获得数组或者切片的一部分
slice := []int{1, 2, 3}     //使用字面量初始化新的切片  //这里的新建编译期会先新建出数组然后走切割
slice := make([]int, 10)    //使用关键字 make 创建切片
```
通过make来新建,运行期新建

runtime.slice
```go
func makeslice64(et *_type, len64, cap64 int64) unsafe.Pointer {
	//这个判断的方式有点骚
	len := int(len64)
	if int64(len) != len64 {
		panicmakeslicelen()
	}
	cap := int(cap64)
	if int64(cap) != cap64 {
		panicmakeslicecap()
	}

	return makeslice(et, len, cap)
}
```
```go
func makeslice(et *_type, len, cap int) unsafe.Pointer {
	// 计算当前切片需要用到的内存空间
	mem, overflow := math.MulUintptr(et.size, uintptr(cap))
	//overflow 溢出 
	// mem > maxAlloc 申请的内容大于最大可申请的内存
	if overflow || mem > maxAlloc || len < 0 || len > cap {
		// 如果cap溢出 或者 大于最大可申请的内存，尝试用len计算
		mem, overflow := math.MulUintptr(et.size, uintptr(len))
		if overflow || mem > maxAlloc || len < 0 {
			panicmakeslicelen()
		}
		panicmakeslicecap()
	}
	return mallocgc(mem, et, true)
}
```
mallocgc 方法
```text
  //分配一个大小为mem字节的对象。 
  // 从per-P缓存的空闲列表中分配小对象。 
  // 从堆直接分配大对象（mem> 32 kB）。
  // ps：看其源码如果mem等于0，直接返回 return unsafe.Pointer(&zerobase)
  // var zerobase uintptr
    mallocgc(mem, et, true)
```
growslice 切片扩容的方法
```go
et 类型
old 旧的切片
cap 所需的最小容量
func growslice(et *_type, old slice, cap int) slice {
	// 开启了竞态检测
	if raceenabled {
		callerpc := getcallerpc()
		racereadrangepc(old.array, uintptr(old.len*int(et.size)), callerpc, funcPC(growslice))
	}
	//是否开启了msan，探测是否读未初始化的内存
	if msanenabled {
		msanread(old.array, uintptr(old.len*int(et.size)))
	}
    // 如果所需容量小于就切片的容量
	if cap < old.cap {
		panic(errorString("growslice: cap out of range"))
	}
    // 如果对应的类型size是0，那就是zerobase 不需要申请内存
	if et.size == 0 {
		return slice{unsafe.Pointer(&zerobase), old.len, cap}
	}
    // 计算需要多少容量
	newcap := old.cap
	doublecap := newcap + newcap
	// 如果所需容量大于old的2倍，那就取所需的,
	// 否则如果old的容量小于1024 取old的2倍
	// 如果old大于1024 每次增加 4分1 直到满足条件
	// 例如 old 1600 need 3000
	//  第一次 2000<3000 第二次 2000+2000/4=2500<3000 第三次2500+2500/4 =3125>3000
	if cap > doublecap {
		newcap = cap
	} else {
		if old.len < 1024 {
			newcap = doublecap
		} else {
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}
			if newcap <= 0 {
				newcap = cap
			}
		}
	}
    
	var overflow bool
	var lenmem, newlenmem, capmem uintptr
	// 计算容量
	// 这里都是细节啊！！
	// 如果size是1 暂时没发现1代表啥哈哈 不需要做乘除的运算
	// 如果size是指针类型的,做相关计算
	// 如果大小是2的倍数的,走了一些计算的优化
	// 其他的不能优化走默认去了
	switch {
	case et.size == 1:
		lenmem = uintptr(old.len)
		newlenmem = uintptr(cap)
		capmem = roundupsize(uintptr(newcap))
		overflow = uintptr(newcap) > maxAlloc
		newcap = int(capmem)
	case et.size == sys.PtrSize:
		lenmem = uintptr(old.len) * sys.PtrSize
		newlenmem = uintptr(cap) * sys.PtrSize
		capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
		overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
		newcap = int(capmem / sys.PtrSize)
	case isPowerOfTwo(et.size):
		var shift uintptr
		if sys.PtrSize == 8 {
			// Mask shift for better code generation.
			shift = uintptr(sys.Ctz64(uint64(et.size))) & 63
		} else {
			shift = uintptr(sys.Ctz32(uint32(et.size))) & 31
		}
		lenmem = uintptr(old.len) << shift
		newlenmem = uintptr(cap) << shift
		capmem = roundupsize(uintptr(newcap) << shift)
		overflow = uintptr(newcap) > (maxAlloc >> shift)
		newcap = int(capmem >> shift)
	default:
		lenmem = uintptr(old.len) * et.size
		newlenmem = uintptr(cap) * et.size
		capmem, overflow = math.MulUintptr(et.size, uintptr(newcap))
		capmem = roundupsize(capmem)
		newcap = int(capmem / et.size)
	}
    // 溢出直接panic
	if overflow || capmem > maxAlloc {
		panic(errorString("growslice: cap out of range"))
	}

	var p unsafe.Pointer
	// 切片不是指针类型调用memclrNoHeapPointers将超出切片当前长度的位置清空，
	// 仅清除不会被覆盖的位置，细节啊...！！！！！
	if et.ptrdata == 0 {
		p = mallocgc(capmem, nil, false)
		// The append() that calls growslice is going to overwrite from old.len to cap (which will be the new length).
		// Only clear the part that will not be overwritten.
		memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
	} else {
		// Note: can't use rawmem (which avoids zeroing of memory), because then GC can scan uninitialized memory.
		p = mallocgc(capmem, et, true)
		if lenmem > 0 && writeBarrier.enabled {
			// Only shade the pointers in old.array since we know the destination slice p
			// only contains nil pointers because it has been cleared during alloc.
			bulkBarrierPreWriteSrcOnly(uintptr(p), uintptr(old.array), lenmem-et.size+et.ptrdata)
		}
	}
	// 旧移动到新
	memmove(p, old.array, lenmem)
    
	return slice{p, old.len, newcap}
}
```
slicecopy: 运行时拷贝切片
```go
toPtr 目标切片
toLen 目标切片长度
fmPtr 来源切片
fmLen 来源切片长度
width 单个元素的大小
func slicecopy(toPtr unsafe.Pointer, toLen int, fmPtr unsafe.Pointer, fmLen int, width uintptr) int {
	if fmLen == 0 || toLen == 0 {
		return 0
	}

	n := fmLen
	if toLen < n {
		n = toLen
	}

	if width == 0 {
		return n
	}

	if raceenabled {
		callerpc := getcallerpc()
		pc := funcPC(slicecopy)
		racereadrangepc(fmPtr, uintptr(n*int(width)), callerpc, pc)
		racewriterangepc(toPtr, uintptr(n*int(width)), callerpc, pc)
	}
	if msanenabled {
		msanread(fmPtr, uintptr(n*int(width)))
		msanwrite(toPtr, uintptr(n*int(width)))
	}

	size := uintptr(n) * width
	// size 等于1指针
	if size == 1 { // common case worth about 2x to do here
		// TODO: is this still worth it with new memmove impl?
		*(*byte)(toPtr) = *(*byte)(fmPtr) // known to be a byte pointer
	} else {
		// 整块内存拷贝
		memmove(toPtr, fmPtr, size)
	}
	return n
}
```