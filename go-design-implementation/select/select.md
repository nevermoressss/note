选择select
```text
让 Goroutine 同时等待多个 Channel 的可读或者可写.
在多个文件或者 Channel 发生状态改变之前，select 会一直阻塞当前线程或者 Goroutine。
select 是一种与 switch 相似的控制结构。
select 中虽然也有多个 case，但是这些 case 中的表达式必须都是 Channel 的收发操作。
```
scase：runtime.scase(select 控制结构中的 case )
```go
type scase struct {
	c           *hchan         // 没有case都会和channal 有关系，所以包含一个channal的地址
	elem        unsafe.Pointer // 接收或者发送数据的变量地址
	kind        uint16         // scase 的种类
	pc          uintptr // race pc (for race detector / msan)
	releasetime int64
}
// 种类
const (
	caseNil = iota
	caseRecv
	caseSend
	caseDefault
)
```
select语句在编译期间会根据case的数量来进行编译器的重写和优化  
cmd/compile/internal/gc.walkselectcases  
1、select 不存在任何的 case；  
2、select 只存在一个 case；  (改写成单channal阻塞)
3、select 存在两个 case，其中一个 case 是 default；  (如果非default的case阻塞了，执行case)
4、select 存在多个 case  
```go
func walkselectcases(cases *Nodes) []*Node {
	n := cases.Len()
	sellineno := lineno
	// 不存在任何case 调用mkcall("block", nil, nil) 
	// 实际会调用 runtime.gopark 让当前G让出对CPU的使用权
	// 直接阻塞当前的 Goroutine，导致 Goroutine 进入无法被唤醒的永久休眠状态。
	if n == 0 {
		return []*Node{mkcall("block", nil, nil)}
	}
	//如果当前的 select 条件只包含一个 case，那么就会将 select 改写成 if 条件语句。
	// 当 case 中的 Channel 是空指针时，就会直接挂起当前 Goroutine 并永久休眠。 
	if n == 1 {
		cas := cases.First()
		setlineno(cas)
		l := cas.Ninit.Slice()
		if cas.Left != nil {
			n := cas.Left
			l = append(l, n.Ninit.Slice()...)
			n.Ninit.Set(nil)
			var ch *Node
			switch n.Op {
			default:
				Fatalf("select %v", n.Op)
			case OSEND:
				ch = n.Left
			case OSELRECV, OSELRECV2:
				ch = n.Right.Left
				if n.Op == OSELRECV || n.List.Len() == 0 {
					if n.Left == nil {
						n = n.Right
					} else {
						n.Op = OAS
					}
					break
				}
				if n.Left == nil {
					nblank = typecheck(nblank, ctxExpr|ctxAssign)
					n.Left = nblank
				}
				n.Op = OAS2
				n.List.Prepend(n.Left)
				n.Rlist.Set1(n.Right)
				n.Right = nil
				n.Left = nil
				n.SetTypecheck(0)
				n = typecheck(n, ctxStmt)
			}
			// 如果ch是空 ,挂起来
			// if ch == nil { block() }; n;
			a := nod(OIF, nil, nil)
			a.Left = nod(OEQ, ch, nodnil())
			var ln Nodes
			ln.Set(l)
			a.Nbody.Set1(mkcall("block", nil, &ln))
			l = ln.Slice()
			a = typecheck(a, ctxStmt)
			l = append(l, a, n)
		}
		l = append(l, cas.Nbody.Slice()...)
		l = append(l, nod(OBREAK, nil, nil))
		return l
	}

	// 通用逻辑
	for _, cas := range cases.Slice() {
		setlineno(cas)
		n := cas.Left
		if n == nil {
			continue
		}
		switch n.Op {
		case OSEND:
			n.Right = nod(OADDR, n.Right, nil)
			n.Right = typecheck(n.Right, ctxExpr)

		case OSELRECV, OSELRECV2:
			if n.Op == OSELRECV2 && n.List.Len() == 0 {
				n.Op = OSELRECV
			}

			if n.Left != nil {
				n.Left = nod(OADDR, n.Left, nil)
				n.Left = typecheck(n.Left, ctxExpr)
			}
		}
	}

	// select 存在两个 case，其中一个 case 是 default；
	if n == 2 && (cases.First().Left == nil || cases.Second().Left == nil) {
		var cas *Node
		var dflt *Node
		if cases.First().Left == nil {
			cas = cases.Second()
			dflt = cases.First()
		} else {
			dflt = cases.Second()
			cas = cases.First()
		}

		n := cas.Left
		setlineno(n)
		r := nod(OIF, nil, nil)
		r.Ninit.Set(cas.Ninit.Slice())
		switch n.Op {
		default:
			Fatalf("select %v", n.Op)

		case OSEND:
			// if selectnbsend(c, v) { body } else { default body }
			ch := n.Left
			r.Left = mkcall1(chanfn("selectnbsend", 2, ch.Type), types.Types[TBOOL], &r.Ninit, ch, n.Right)

		case OSELRECV:
			// if selectnbrecv(&v, c) { body } else { default body }
			r = nod(OIF, nil, nil)
			r.Ninit.Set(cas.Ninit.Slice())
			ch := n.Right.Left
			elem := n.Left
			if elem == nil {
				elem = nodnil()
			}
			r.Left = mkcall1(chanfn("selectnbrecv", 2, ch.Type), types.Types[TBOOL], &r.Ninit, elem, ch)

		case OSELRECV2:
			// if selectnbrecv2(&v, &received, c) { body } else { default body }
			r = nod(OIF, nil, nil)
			r.Ninit.Set(cas.Ninit.Slice())
			ch := n.Right.Left
			elem := n.Left
			if elem == nil {
				elem = nodnil()
			}
			receivedp := nod(OADDR, n.List.First(), nil)
			receivedp = typecheck(receivedp, ctxExpr)
			r.Left = mkcall1(chanfn("selectnbrecv2", 2, ch.Type), types.Types[TBOOL], &r.Ninit, elem, receivedp, ch)
		}

		r.Left = typecheck(r.Left, ctxExpr)
		r.Nbody.Set(cas.Nbody.Slice())
		r.Rlist.Set(append(dflt.Ninit.Slice(), dflt.Nbody.Slice()...))
		return []*Node{r, nod(OBREAK, nil, nil)}
	}

	var init []*Node

	// generate sel-struct
	lineno = sellineno
	selv := temp(types.NewArray(scasetype(), int64(n)))
	r := nod(OAS, selv, nil)
	r = typecheck(r, ctxStmt)
	init = append(init, r)

	order := temp(types.NewArray(types.Types[TUINT16], 2*int64(n)))
	r = nod(OAS, order, nil)
	r = typecheck(r, ctxStmt)
	init = append(init, r)

	// register cases
	for i, cas := range cases.Slice() {
		setlineno(cas)

		init = append(init, cas.Ninit.Slice()...)
		cas.Ninit.Set(nil)

		// Keep in sync with runtime/select.go.
		const (
			caseNil = iota
			caseRecv
			caseSend
			caseDefault
		)

		var c, elem *Node
		var kind int64 = caseDefault

		if n := cas.Left; n != nil {
			init = append(init, n.Ninit.Slice()...)

			switch n.Op {
			default:
				Fatalf("select %v", n.Op)
			case OSEND:
				kind = caseSend
				c = n.Left
				elem = n.Right
			case OSELRECV, OSELRECV2:
				kind = caseRecv
				c = n.Right.Left
				elem = n.Left
			}
		}

		setField := func(f string, val *Node) {
			r := nod(OAS, nodSym(ODOT, nod(OINDEX, selv, nodintconst(int64(i))), lookup(f)), val)
			r = typecheck(r, ctxStmt)
			init = append(init, r)
		}

		setField("kind", nodintconst(kind))
		if c != nil {
			c = convnop(c, types.Types[TUNSAFEPTR])
			setField("c", c)
		}
		if elem != nil {
			elem = convnop(elem, types.Types[TUNSAFEPTR])
			setField("elem", elem)
		}

		// TODO(mdempsky): There should be a cleaner way to
		// handle this.
		if instrumenting {
			r = mkcall("selectsetpc", nil, nil, bytePtrToIndex(selv, int64(i)))
			init = append(init, r)
		}
	}

	// run the select
	lineno = sellineno
	chosen := temp(types.Types[TINT])
	recvOK := temp(types.Types[TBOOL])
	r = nod(OAS2, nil, nil)
	r.List.Set2(chosen, recvOK)
	fn := syslook("selectgo")
	r.Rlist.Set1(mkcall1(fn, fn.Type.Results(), nil, bytePtrToIndex(selv, 0), bytePtrToIndex(order, 0), nodintconst(int64(n))))
	r = typecheck(r, ctxStmt)
	init = append(init, r)

	// selv and order are no longer alive after selectgo.
	init = append(init, nod(OVARKILL, selv, nil))
	init = append(init, nod(OVARKILL, order, nil))

	// dispatch cases
	for i, cas := range cases.Slice() {
		setlineno(cas)

		cond := nod(OEQ, chosen, nodintconst(int64(i)))
		cond = typecheck(cond, ctxExpr)
		cond = defaultlit(cond, nil)

		r = nod(OIF, cond, nil)

		if n := cas.Left; n != nil && n.Op == OSELRECV2 {
			x := nod(OAS, n.List.First(), recvOK)
			x = typecheck(x, ctxStmt)
			r.Nbody.Append(x)
		}

		r.Nbody.AppendNodes(&cas.Nbody)
		r.Nbody.Append(nod(OBREAK, nil, nil))
		init = append(init, r)
	}

	return init
}
```
select 流程
```text
ps :参考 https://www.jianshu.com/p/17527019285e
将所有的 case 转换成包含 Channel 以及类型等信息的 runtime.scase 结构体；  
调用运行时函数 runtime.selectgo 从多个准备就绪的 Channel 中选择一个可执行的 runtime.scase 结构体；  
通过 for 循环生成一组 if 语句，在语句中判断自己是不是被选中的 case  
1. 锁定scase语句中所有的channel

2. 按照随机顺序检测scase中的channel是否ready

　　2.1 如果case可读，则读取channel中数据，解锁所有的channel，然后返回(case index)

　　2.2 如果case可写，则将数据写入channel，解锁所有的channel，然后返回(case index)

　　2.3 所有case都未ready，则解锁所有的channel，然后返回（default index）

3. 所有case都未ready，且没有default语句

　　 3.1 将当前协程加入到所有channel的等待队列

 　　3.2 当将协程转入阻塞，等待被唤醒

4. 唤醒后返回channel对应的case index

　　4.1 如果是读操作，解锁所有的channel，然后返回(case index)

　　4.2 如果是写操作，解锁所有的channel，然后返回(case index)
```