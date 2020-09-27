todo 
#### 反射
```text
 Go 语言的 reflect 包能够弥补它在语法上的一些劣势。
 Go 中一个元素包含类型和值两部分
```
reflect.TypeOf 获取类型
```text
func TypeOf(i interface{}) Type {
    // 转成一个基础的emptyInterface获取类型
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	return toType(eface.typ)
}
```
reflect.ValueOf 获取类型
```text
// ValueOf返回一个初始化为具体值的新Value
func ValueOf(i interface{}) Value {
	if i == nil {
		return Value{}
	}
	// 在堆上分配内存
	escapes(i)

	return unpackEface(i)
}
```
Elem()方法
```text
获取指针指向的变量；
```
reflect.Value.Set() 方法
```text
调用 reflect.Value.assignTo 并返回一个新的反射对象，这个返回的反射对象指针就会直接覆盖原始的反射变量。
```
type 和value 还提供了能多其他的方法，都是为了能够对传进来的interface做一系列操作，就不一一举例啦.. 