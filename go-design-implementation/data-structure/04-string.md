##### string
```text
Go 语言中的字符串其实是一个只读的字节数组
```
string 对应的结构
```go
type StringHeader struct {
	Data uintptr    //指向底层数据
	Len  int        //长度
}

type stringStruct struct {
	str unsafe.Pointer
	len int
}

```
字符串拼接
concatstrings：runtime/concatstrings
```go
func concatstrings(buf *tmpBuf, a []string) string {
	idx := 0
	l := 0
	count := 0
	// 先把空的过滤了
	for i, x := range a {
		n := len(x)
		if n == 0 {
			continue
		}
		if l+n < l {
			throw("string concatenation too long")
		}
		l += n
		count++
		idx = i
	}
	if count == 0 {
		return ""
	}
	// 如果非空字符串的数量为 1 并且当前的字符串不在栈上就可以直接返回该字符串，不需要进行额外的任何操作
	if count == 1 && (buf != nil || !stringDataOnStack(a[idx])) {
		return a[idx]
	}
	// 申请新的内存
	s, b := rawstringtmp(buf, l)
	// 多个字符串拷贝到目标字符串所在的内存空间中
	for _, x := range a {
		copy(b, x)
		b = b[len(x):]
	}
	return s
}
```
字节数组到字符串的转换  
slicebytetostring：runtime/slicebytetostring
```go
func slicebytetostring(buf *tmpBuf, ptr *byte, n int) (str string) {
	if n == 0 {
		// 长度为0特殊处理
		return ""
	}
	// 长度为1 也特殊处理
	if n == 1 {
		p := unsafe.Pointer(&staticuint64s[*ptr])
		if sys.BigEndian {
			p = add(p, 7)
		}
		stringStructOf(&str).str = p
		stringStructOf(&str).len = 1
		return
	}

	var p unsafe.Pointer
	// 如果入参的缓冲区够用
	if buf != nil && n <= len(buf) {
		p = unsafe.Pointer(buf)
	} else {
		// 申请内存
		p = mallocgc(uintptr(n), nil, false)
	}
	// 复制
	stringStructOf(&str).str = p
	stringStructOf(&str).len = n
	// 移动
	memmove(p, unsafe.Pointer(ptr), uintptr(n))
	return
}
```
字符串转字节数组
stringtoslicebyte：runtime/stringtoslicebyte
```go
func stringtoslicebyte(buf *tmpBuf, s string) []byte {
	var b []byte
	// 缓冲区够用
	if buf != nil && len(s) <= len(buf) {
		*buf = tmpBuf{}
		b = buf[:len(s)]
	} else {
		// 不够就申请内存
		b = rawbyteslice(len(s))
	}
	copy(b, s)
	return b
}
```