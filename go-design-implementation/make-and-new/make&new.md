#### make&new
```text
make 的作用是初始化内置的数据结构，切片、哈希表和 Channel
new 的作用是根据传入的类型分配一片内存空间并返回指向这片内存空间的指针(返回的是指针)

Go 语言就将代表 make 关键字的 OMAKE 节点根据参数类型的不同转换成了 OMAKESLICE、OMAKEMAP 和 OMAKECHAN 三种不同类型的节点，这些节点会调用不同的运行时函数来初始化相应的数据结构。

new 如果申请的内存为0，返回zerobase ，其他则在编译期转换成 runtime.newobject
```
newobject：runtime.newobject
```go
// implementation of new builtin
// compiler (both frontend and SSA backend) knows the signature
// of this function
// 获取对应大小的内存
func newobject(typ *_type) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}
```