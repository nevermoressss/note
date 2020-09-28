#### sync 同步
```text
sync.Mutex
sync.RWMutex
sync.WaitGroup
sync.Once
sync.Cond
```
#### Mutex
```go
// A Mutex must not be copied after first use.
type Mutex struct {
	state int32
	sema  uint32
}
```
mutex.Lock()
```go
func (m *Mutex) Lock() {
	// 先看看能不能直接获取到
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		// 能就直接返回
		return
	}
	// 不能直接获取到就调用lockSlow 慢慢获取
	m.lockSlow()
}
```
lockSlow()
```go
func (m *Mutex) lockSlow() {
	var waitStartTime int64
	starving := false
	awoke := false
	iter := 0
	old := m.state
	for {
		// 不是饥饿模式下尝试通过自旋获取
		//runtime_canSpin为true的条件
		//1运行在多 CPU 的机器上
		//2当前 Goroutine 为了获取该锁进入自旋的次数小于四次
		//3当前机器上至少存在一个正在运行的处理器 P 并且处理的运行队列为空
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
			runtime_doSpin()
			iter++
			old = m.state
			continue
		}
		new := old
		// Don't try to acquire starving mutex, new arriving goroutines must queue.
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}
		// 切换为饥饿模式
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}
		if awoke {
			// 被唤醒
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}
		//通过cas获取锁
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			if old&(mutexLocked|mutexStarving) == 0 {
				break // locked the mutex with CAS
			}
			// 如果之前已经在等待，就在队列的最前面等待.
			queueLifo := waitStartTime != 0
			if waitStartTime == 0 {
				waitStartTime = runtime_nanotime()
			}
			// 休眠
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
			old = m.state
			if old&mutexStarving != 0 {
				// 如果被唤醒了并且互斥锁处于饥饿状态，那么我们就是饥饿模式下的管理者
				if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
					throw("sync: inconsistent mutex state")
				}
				delta := int32(mutexLocked - 1<<mutexWaiterShift)
				if !starving || old>>mutexWaiterShift == 1 {
					// 退出饥饿模式
					delta -= mutexStarving
				}
				atomic.AddInt32(&m.state, delta)
				break
			}
			awoke = true
			iter = 0
		} else {
			old = m.state
		}
	}
	if race.Enabled {
		race.Acquire(unsafe.Pointer(m))
	}
}
```
```text
如果互斥锁处于初始化状态，就会直接通过置位 mutexLocked 加锁；

如果互斥锁处于 mutexLocked 并且在普通模式下工作，就会进入自旋，执行 30 次 PAUSE 指令消耗 CPU 时间等待锁的释放；

如果当前 Goroutine 等待锁的时间超过了 1ms，互斥锁就会切换到饥饿模式；

互斥锁在正常情况下会通过 sync.runtime_SemacquireMutex 函数将尝试获取锁的 Goroutine 切换至休眠状态，等待锁的持有者唤醒当前 Goroutine；

如果当前 Goroutine 是互斥锁上的最后一个等待的协程或者等待的时间小于 1ms，当前 Goroutine 会将互斥锁切换回正常模式；
```
mutex.UnLock()
```go
func (m *Mutex) Unlock() {
	//如果该函数返回的新状态等于 0，当前 Goroutine 就成功解锁了互斥锁；
    //如果该函数返回的新状态不等于 0，这段代码会调用 sync.Mutex.unlockSlow 方法开始慢速解锁：
	if race.Enabled {
		_ = m.state
		race.Release(unsafe.Pointer(m))
	}
	// 快速解锁
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		//在正常模式下如果互斥锁不存在等待者或者互斥锁的 mutexLocked、mutexStarving、mutexWoken 状态不都为 0，那么当前方法就可以直接返回，不需要唤醒其他等待者；
		//在正常模式下如果互斥锁存在等待者，会通过 sync.runtime_Semrelease 唤醒等待者并移交锁的所有权；
		//在饥饿模式下，代码会直接调用 sync.runtime_Semrelease 方法将当前锁交给下一个正在尝试获取锁的等待者，等待者被唤醒后会得到锁，在这时互斥锁还不会退出饥饿状态；
		m.unlockSlow(new)
	}
}
```
#### RWMutex
```go
type RWMutex struct {
	w           Mutex   // 复用mutex
	writerSem   uint32  // 写等待读
	readerSem   uint32  // 读等待写
	readerCount int32   // 正在执行的读操作个数
	readerWait  int32   // 写操作阻塞了之后读操作的等待个数
}
```
RWMutex.Lock()
```go
func (rw *RWMutex) Lock() {
	// 锁上 只能一个写进入
	rw.w.Lock()
	// 标记 有写在等待
	r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
	// 等待之前的读操作执行完
	if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
		runtime_SemacquireMutex(&rw.writerSem, false, 0)
	}
}
```
RWMutex.UnLock()
```go
func (rw *RWMutex) Unlock() {
	// 可以执行读操作了
	r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
	if r >= rwmutexMaxReaders {
		race.Enable()
		throw("sync: Unlock of unlocked RWMutex")
	}
	// 通过 for 循环触发所有由于获取读锁而陷入等待的 Goroutine：
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
	// 释放锁
	rw.w.Unlock()
}
```
RWMutex.RLock()
```go
func (rw *RWMutex) RLock() {
	//如果该方法返回负数 — 其他 Goroutine 获得了写锁，当前 Goroutine 就会调用 sync.runtime_SemacquireMutex 陷入休眠等待锁的释放；
        //如果该方法的结果为非负数 — 没有 Goroutine 获得写锁，当前方法就会成功返回；
	if atomic.AddInt32(&rw.readerCount, 1) < 0 {
		// A writer is pending, wait for it.
		runtime_SemacquireMutex(&rw.readerSem, false, 0)
	}
}
```
RWMutex.RUnlock()
```go
func (rw *RWMutex) RUnlock() {
    //如果返回值大于等于零 — 读锁直接解锁成功；
    //如果返回值小于零 — 有一个正在执行的写操作，在这时会调用sync.RWMutex.rUnlockSlow 方法；
    //sync.RWMutex.rUnlockSlow 会减少获取锁的写操作等待的读操作数 readerWait 并在所有读操作都被释放之后触发写操作的信号量 writerSem，该信号量被触发时，调度器就会唤醒尝试获取写锁的 Goroutine。
	if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
		// Outlined slow-path to allow the fast-path to be inlined
		rw.rUnlockSlow(r)
	}
}
```
#### WaitGroup
```go
type WaitGroup struct {
	// 不能通过再拷贝的方式赋值
	noCopy noCopy
	//状态和信号量
	state1 [3]uint32
}
```
WaitGroup.Add()
```go
func (wg *WaitGroup) Add(delta int) {
	// 获取状态和信号量，因为state1 在32和64系统里代表不一样的，所以有个方法来获取状态和信号量
	statep, semap := wg.state()
	// 增加
	state := atomic.AddUint64(statep, uint64(delta)<<32)
	v := int32(state >> 32)
	w := uint32(state)
	// 总数小于0 panic
	if v < 0 {
		panic("sync: negative WaitGroup counter")
	}
	// 
	if w != 0 && delta > 0 && v == int32(delta) {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	if v > 0 || w == 0 {
		return
	}
	if *statep != state {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	// Reset waiters count to 0.
	*statep = 0
	for ; w != 0; w-- {
		// 唤醒
		runtime_Semrelease(semap, false, 0)
	}
}
```
WaitGroup.Wait()
```go
// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	// 获取状态和信号量
	statep, semap := wg.state()
	// 
	for {
		// 查看当前的数量
		state := atomic.LoadUint64(statep)
		v := int32(state >> 32)
		w := uint32(state)
		// 为0直接返回
		if v == 0 {
			// Counter is 0, no need to wait.
			return
		}
		// 记录等待
		if atomic.CompareAndSwapUint64(statep, state, state+1) {
			// 挂起 等调用add方法到0唤醒
			runtime_Semacquire(semap)
			if *statep != 0 {
				panic("sync: WaitGroup is reused before previous Wait has returned")
			}
			if race.Enabled {
				race.Enable()
				race.Acquire(unsafe.Pointer(wg))
			}
			return
		}
	}
}
```
#### Once
保证在 Go 程序运行期间的某段代码只会执行一次。
```go
type Once struct {
	// 标记是否已经执行
	done uint32
	m    Mutex
}
```
Once.Do()
```go
func (o *Once) Do(f func()) {
	// 如果done为0
	if atomic.LoadUint32(&o.done) == 0 {
		// Outlined slow-path to allow inlining of the fast-path.
		o.doSlow(f)
	}
}
// 保证只执行一次
func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}
```
#### Once.Cond()
让一系列的 Goroutine 都在满足特定条件时被唤醒。
```go
type Cond struct {
	noCopy  noCopy  // 编译器不允许拷贝
	L       Locker
	notify  notifyList
	checker copyChecker //运行期不允许拷贝
}
type notifyList struct {
	wait   uint32           //正在等待的
	notify uint32           //已经通知的
	lock   uintptr // key field of the mutex
	head   unsafe.Pointer   //链表头
	tail   unsafe.Pointer   //链表尾
}
```
Cond.Wait()
```go
// wait方法一看 调用wait前要加锁，wait后要解锁，不然就panic啦
func (c *Cond) Wait() {
	c.checker.check()
	t := runtime_notifyListAdd(&c.notify)   // 等待+1
	c.L.Unlock()                        // 解锁
	runtime_notifyListWait(&c.notify, t)    //挂起等待
	c.L.Lock() //加锁
}
```
Cond.Signal() 唤醒一个在wait中的（队头）
```go
func (c *Cond) Signal() {
	c.checker.check()
	runtime_notifyListNotifyOne(&c.notify)
}
// 找到一个满足条件的唤醒
func notifyListNotifyOne(l *notifyList) {
	t := l.notify
	atomic.Store(&l.notify, t+1)
	for p, s := (*sudog)(nil), l.head; s != nil; p, s = s, s.next {
		if s.ticket == t {
			n := s.next
			if p != nil {
				p.next = n
			} else {
				l.head = n
			}
			if n == nil {
				l.tail = p
			}
			s.next = nil
			readyWithTime(s, 4)
			return
		}
	}
}
```
Cond.Broadcast() 唤醒全部
```go
func (c *Cond) Broadcast() {
	c.checker.check()
	runtime_notifyListNotifyAll(&c.notify)
}
// 唤醒全部
func notifyListNotifyAll(l *notifyList) {
	s := l.head
	l.head = nil
	l.tail = nil

	atomic.Store(&l.notify, atomic.Load(&l.wait))

	for s != nil {
		next := s.next
		s.next = nil
		readyWithTime(s, 4)
		s = next
	}
}
```