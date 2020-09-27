#### 概述
接口的本质就是引入一个新的中间层，调用方可以通过接口与具体实现分离，解除上下游的耦合，上层的模块不再需要依赖下层的具体模块，只需要依赖一个约定好的接口。  

在 Go 中：实现接口的所有方法就隐式的实现了接口；Go 语言中接口的实现都是隐式的

接口也是 Go 语言中的一种类型，它能够出现在变量的定义、函数的入参和返回值中并对它们进行约束，不过 Go 语言中有两种略微不同的接口，一种是带有一组方法的接口，另一种是不带任何方法的 interface{}

不包含任何方法的 interface{} 类型  
eface:runtime/eface
```go
type eface struct { // 16 bytes
	_type *_type            //类型
	data  unsafe.Pointer    //指针
}
//只包含指向底层数据和类型的两个指针，所以任意类型都可以转换成 interface{} 类型
```
包含方法的类型 
```go
type iface struct {
	tab  *itab              //接口类型和具体类型的组合
	data unsafe.Pointer     //指针
}

type itab struct {
	inter *interfacetype
	_type *_type        //类型
	hash  uint32 // _type.hash 里拷贝出来的， 用于快速判断类型是否相等.
	_     [4]byte       
	fun   [1]uintptr // 函数指针.
}
```