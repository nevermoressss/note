#### context
上下文 context.Context 是用来设置截止日期、同步信号，传递请求相关值的结构体。

该接口定义了四个需要实现的方法

Deadline — 返回 context.Context 被取消的时间，也就是完成工作的截止日期；

Done — 返回一个 Channel，这个 Channel 会在当前工作完成或者上下文被取消之后关闭，多次调用 Done 方法会返回同一个 Channel；

Err — 返回 context.Context 结束的原因，它只会在 Done 返回的 Channel 被关闭时才会返回非空的值；

1 如果 context.Context 被取消，会返回 Canceled 错误；

2 如果 context.Context 超时，会返回 DeadlineExceeded 错误；

Value — 从 context.Context中获取键对应的值，对于同一个上下文来说，多次调用 Value 并传入相同的 Key会返回相同的结果，该方法可以用来传递请求特定的数据；
```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}
```
###### 默认ctx（context.Background、context.TODO）
```go
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}
```

###### context.WithCancel
```go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := newCancelCtx(parent) //将parent包装成cancelctx
	propagateCancel(parent, &c) //关联
	return &c, func() { c.cancel(true, Canceled) }
}

func propagateCancel(parent Context, child canceler) {
    //当parent的done为空，也就是说parent不会触发done事件，那么child也肯定不会done直接return
	if parent.Done() == nil {
		return // parent is never canceled
	}
    //判断parent的类型，是不是cancelCtx和timerCtx
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		//parent 的 err不为空，那就把child的关了
		if p.err != nil {
			// parent has already been canceled
			child.cancel(false, p.err)
		} else {
		    //如果parent的children为空，还没初始化map
			if p.children == nil {
				p.children = make(map[canceler]struct{})
			}
			p.children[child] = struct{}{}
		}
		p.mu.Unlock()
	} else {
	    //parentcontext done了 调用child的cancel方法，child done了和parent无关
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}
```

###### context.WithDeadline 和 context.WithTimeout

```go
//WithTimeout把传入的时间转成准确时间调用WithDeadline方法
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
    //parent 的截止日期比传入的早，以parent的为主,return
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		// The current deadline is already sooner than the new one.
		return WithCancel(parent)
	}
	//创建timectx
	c := &timerCtx{
		cancelCtx: newCancelCtx(parent),
		deadline:  d,
	}
	propagateCancel(parent, c)
	dur := time.Until(d)
	if dur <= 0 {
		c.cancel(true, DeadlineExceeded) // deadline has already passed
		return c, func() { c.cancel(false, Canceled) }
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	//创建定时器
	if c.err == nil {
		c.timer = time.AfterFunc(dur, func() {
			c.cancel(true, DeadlineExceeded)
		})
	}
	return c, func() { c.cancel(true, Canceled) }
}
```