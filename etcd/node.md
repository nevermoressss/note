node:对rawnode安全性封装
```text
// 可以看到node 对于rawNode多了一系列的chan
type node struct {
    // 用于接收MsgProp消息
	propc      chan msgWithResult
	// 除MsgProp外的其他消息
	recvc      chan pb.Message
	// 更改配置消息
	confc      chan pb.ConfChangeV2
	// 当前集群中所有节点的ID
	confstatec chan pb.ConfState
	// 用于向上层模块返回ready
	readyc     chan Ready
	// 用于通知底层的raft模块 ready数据已处理完
	advancec   chan struct{}
	// 接收逻辑时钟的信号
	tickc      chan struct{}
	// 
	done       chan struct{}
	stop       chan struct{}
	status     chan chan Status
    // 
	rn *RawNode
}
```
Raedy 结构：用于传递数据，字段都是可读的
```text
type Ready struct {
	// 当前主节点id 和 当前节点角色
	*SoftState
    // 当前任期 、投票给了什么节点、当前节点已提交的位置
	pb.HardState
    // 当前节点中等待处理的只读请求
   	ReadStates []ReadState
    // 需要上层模块保存在storage中
	Entries []pb.Entry
    // 持久化的快照数据
    Snapshot pb.Snapshot
    // 指定要提交给store/state-machine的额数据
	CommittedEntries []pb.Entry
    // 当前节点中等待发送到集群其他节点的消息
	Messages []pb.Message
    // 是否需要将hardState 和 entries 同步写入磁盘
    MustSync bool
}
```
newReady 方法
```text
func newReady(r *raft, prevSoftSt *SoftState, prevHardSt pb.HardState) Ready {
	rd := Ready{
	    // unstable部分记录的entries
		Entries:          r.raftLog.unstableEntries(),
		// 已提交未应用的commit
		CommittedEntries: r.raftLog.nextEnts(),
		// 待发送的消息
		Messages:         r.msgs,
	}
	// 判断softState是否有更改，有就返回相关更改
	if softSt := r.softState(); !softSt.equal(prevSoftSt) {
		rd.SoftState = softSt
	}
	// 判断hardState是否有更改，有就返回相关更改
	if hardSt := r.hardState(); !isHardStateEqual(hardSt, prevHardSt) {
		rd.HardState = hardSt
	}
	// 是否记录了新的快照
	if r.raftLog.unstable.snapshot != nil {
		rd.Snapshot = *r.raftLog.unstable.snapshot
	}
	// 只读请求
	if len(r.readStates) != 0 {
		rd.ReadStates = r.readStates
	}
	// 如果 Raft 条目的硬状态和计数表明需要同步写入持久存储，则 MustSync 返回 true。
	rd.MustSync = MustSync(r.hardState(), prevHardSt, len(rd.Entries))
	return rd
}

```
node 方法
```text
// 开启note
func StartNode(c *Config, peers []Peer) Node {
	if len(peers) == 0 {
		panic("no peers given; use RestartNode instead")
	}
	// 初始化RawNode
	rn, err := NewRawNode(c)
	if err != nil {
		panic(err)
	}
	// 引导程序
	rn.Bootstrap(peers)
    // 构建node
	n := newNode(rn)
    // 启用协程处理一系列操作
	go n.run()
	return &n
}
```