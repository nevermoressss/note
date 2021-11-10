rawnode :对raft结构的上层封装
```text
RawNode 是一个线程不安全的节点。该结构体的方法对应于 Node 的方法，并在那里进行了更全面的描述
type RawNode struct {
	raft       *raft    // raft模块
	prevSoftSt *SoftState   // 包含leader节点id 和 节点状态
	prevHardSt pb.HardState // 包含任期、本轮投票给谁、本节点已提交log的位置
}
```
NewRawNode：初始化
```text
func NewRawNode(config *Config) (*RawNode, error) {
    // 初始化raft模块
	r := newRaft(config)
	// 组装rawNode
	rn := &RawNode{
		raft: r,
	}
	rn.prevSoftSt = r.softState()
	rn.prevHardSt = r.hardState()
	return rn, nil
}
```
rawNode 方法介绍
```text
tick:逻辑时间偏移处理函数 对应 raft模块的tickElection或者tickHeartbeat方法
func (rn *RawNode) Tick() {
	rn.raft.tick()
}

TickQuiesced:TickQuiesced 将内部逻辑时钟提前一个滴答，而不执行任何其他状态机处理。 当已知 Raft 组中的所有对等点处于相同状态时，它允许调用者避免周期性心跳和选举。
ps:该方法没有应用的地方
func (rn *RawNode) TickQuiesced() {
	rn.raft.electionElapsed++
}

Campaign：转换为候选节点 PreCandidate或者Candidate
func (rn *RawNode) Campaign() error {
	return rn.raft.Step(pb.Message{
		Type: pb.MsgHup,
	})
}

Propose：将数据附加到raft log
func (rn *RawNode) Propose(data []byte) error {
	return rn.raft.Step(pb.Message{
		Type: pb.MsgProp,
		From: rn.raft.id,
		Entries: []pb.Entry{
			{Data: data},
		}})
}

ApplyConfChange：配置更改
func (rn *RawNode) ApplyConfChange(cc pb.ConfChangeI) *pb.ConfState {
	cs := rn.raft.applyConfChange(cc.AsV2())
	return &cs
}

Step：状态推进
func (rn *RawNode) Step(m pb.Message) error {
	// ignore unexpected local messages receiving over network
	if IsLocalMsg(m.Type) {
		return ErrStepLocalMsg
	}
	if pr := rn.raft.prs.Progress[m.From]; pr != nil || !IsResponseMsg(m.Type) {
		return rn.raft.Step(m)
	}
	return ErrStepPeerNotFound
}

ReportUnreachable:报告对应的节点无法到达
func (rn *RawNode) ReportUnreachable(id uint64) {
	_ = rn.raft.Step(pb.Message{Type: pb.MsgUnreachable, From: id})
}

ReportSnapshot：报告发送快照的状态
func (rn *RawNode) ReportSnapshot(id uint64, status SnapshotStatus) {
	rej := status == SnapshotFailure

	_ = rn.raft.Step(pb.Message{Type: pb.MsgSnapStatus, From: id, Reject: rej})
}

TransferLeader:尝试做领导者转移
func (rn *RawNode) TransferLeader(transferee uint64) {
	_ = rn.raft.Step(pb.Message{Type: pb.MsgTransferLeader, From: transferee})
}
```


