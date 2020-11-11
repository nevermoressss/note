##### raft 结构体
```go
type raft struct {
	id uint64	//集群中的标识

	Term uint64 	//当前任期
	Vote uint64		//当前任期投票给了谁

	readStates []ReadState
	// the log
	raftLog *raftLog  //本地log

	maxMsgSize         uint64 //单条消息最大字节
	maxUncommittedSize uint64   //最大为commit数量
	// TODO(tbg): rename to trk.
	prs tracker.ProgressTracker //其他节点的相关信息

	state StateType //当前节点在集群中的角色

	// isLearner is true if the local raft node is a learner.
	isLearner bool

	msgs []pb.Message //当前节点等待发送的消息

	// the leader id
	lead uint64 //leader节点的id
	// leadTransferee is id of the leader transfer target when its value is not zero.
	// Follow the procedure defined in raft thesis 3.10.
	leadTransferee uint64   //leader转移的目标

	pendingConfIndex uint64 //用于判断可否修改配置

	uncommittedSize uint64  //当前uncommit个数

	readOnly *readOnly

	// number of ticks since it reached last electionTimeout when it is leader
	// or candidate.
	// number of ticks since it reached last electionTimeout or received a
	// valid message from current leader when it is a follower.
	electionElapsed int //选举计时器的时间

	// number of ticks since it reached last heartbeatTimeout.  
	// only leader keeps heartbeatElapsed.
	heartbeatElapsed int //心跳计时器 

	checkQuorum bool
	preVote     bool

	heartbeatTimeout int //心跳超时时间
	electionTimeout  int //选举计时器超时时间
	// randomizedElectionTimeout is a random number between
	// [electiontimeout, 2 * electiontimeout - 1]. It gets reset
	// when raft changes its state to follower or candidate.
	randomizedElectionTimeout int //选举计时器的上限
	disableProposalForwarding bool
    //当前节点推进逻辑时钟的函数，leader和follower、candidate不同
	tick func() 
	step stepFunc
	//当前节点收到消息的处理函数leader、follower、candidate都不同

	logger Logger
}
```
##### config
config结构体,创建raft实例需要用到的参数
```go
// Config contains the parameters to start a raft.
type Config struct {
	// 本地节点的id，不能为0
	ID uint64
	//集群中其他各节点的id
	peers []uint64
	// learners 节点不参与投票，直到追赶上leader节点
	learners []uint64
	// 用于初始化raft的electionTimeout，逻辑时钟推送多少次后触发选举
	ElectionTick int
	// 初始化，raft的heartbeatTimeout，leader节点触发心跳的时长
	HeartbeatTick int
	//日志存储
	Storage Storage
	// 当前已经应用的记录位置
	Applied uint64
	// 初始化maxMsgSize，每条消息的最大字节数
	MaxSizePerMsg uint64
	// 可以应用的已提交的条目大小
	MaxCommittedSizePerReady uint64
	// leader未提交的总字节大小
	MaxUncommittedEntriesSize uint64
	// 初始化maxInflight已经发出去且未收到应答的最大消息个数
	MaxInflightMsgs int

	// 是否开启checkQuorum机制
	CheckQuorum bool

	//是否开启preVote机制
	PreVote bool
	// 指定如何处理只读请求
	ReadOnlyOption ReadOnlyOption

	Logger Logger

	DisableProposalForwarding bool
}
```
reset
```go
func (r *raft) reset(term uint64) {
    //重置
	if r.Term != term {
		r.Term = term
		r.Vote = None
	}
	//lead置空
	r.lead = None
    //重置选举和心跳计时器
	r.electionElapsed = 0
	r.heartbeatElapsed = 0
	//过期时间随机值
	r.resetRandomizedElectionTimeout()
    //清空
	r.abortLeaderTransfer()
    //重置
	r.prs.ResetVotes()
	r.prs.Visit(func(id uint64, pr *tracker.Progress) {
		*pr = tracker.Progress{
			Match:     0,
			Next:      r.raftLog.lastIndex() + 1,
			Inflights: tracker.NewInflights(r.prs.MaxInflight),
			IsLearner: pr.IsLearner,
		}
		if id == r.id {
			pr.Match = r.raftLog.lastIndex()
		}
	})

	r.pendingConfIndex = 0
	r.uncommittedSize = 0
	r.readOnly = newReadOnly(r.readOnly.option)
}
```
Candidate和Follower的定时推进electionElapsed并判断超时了没
```go
// tickElection is run by followers and candidates after r.electionTimeout.
func (r *raft) tickElection() {
	r.electionElapsed++

	if r.promotable() && r.pastElectionTimeout() {
		r.electionElapsed = 0
		r.Step(pb.Message{From: r.id, Type: pb.MsgHup})
	}
}
```
如果超时调用becomePreCandidate（）成为pre  
获得半数节点响应调用Candidate（）成为选举者  
获得半数以上becomeLeader（）成为leader了啊
