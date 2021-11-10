node:对rawnode安全性封装
```text
// 可以看到node 对于rawNode多了一系列的chan
type node struct {
	propc      chan msgWithResult
	recvc      chan pb.Message
	confc      chan pb.ConfChangeV2
	confstatec chan pb.ConfState
	readyc     chan Ready
	advancec   chan struct{}
	tickc      chan struct{}
	done       chan struct{}
	stop       chan struct{}
	status     chan chan Status

	rn *RawNode
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