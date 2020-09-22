```text
Go 语言中数组在初始化之后大小就无法改变，
存储元素类型相同、但是大小不同的数组类型在 Go 语言看来也是完全不同的，
只有两个条件都相同才是同一个类型。
```
#####数组初始化
```go
// cmd/compile/internal/types/type.NewArray
// NewArray returns a new fixed-length array Type.
// param elem ：数组的类型
// param bound ： 数组的长度
func NewArray(elem *Type, bound int64) *Type {
	if bound < 0 {
		Fatalf("NewArray: invalid bound %v", bound)
	}
	t := New(TARRAY)
	t.Extra = &Array{Elem: elem, Bound: bound}
	t.SetNotInHeap(elem.NotInHeap())
	return t
}
```