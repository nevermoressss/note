##### MemoryStorage
在内存上维护状态信息、快照数据、Entry
```go
type MemoryStorage struct {
	// Protects access to all fields. Most methods of MemoryStorage are
	// run on the raft goroutine, but Append() is run on an application
	// goroutine.
	sync.Mutex
    //状态信息（当前任期，当前节点投票给了谁，已提交的entry记录的位置）
	hardState pb.HardState
	//快照信息
	snapshot  pb.Snapshot
	// ents[i] has raft log position i+snapshot.Metadata.Index
	//快照数据之后的所有entry记录
	ents []pb.Entry
}

type HardState struct {
    //当前任期
	Term             uint64 `protobuf:"varint,1,opt,name=term" json:"term"`
	//投票给了谁
	Vote             uint64 `protobuf:"varint,2,opt,name=vote" json:"vote"`
	//已提交的位置
	Commit           uint64 `protobuf:"varint,3,opt,name=commit" json:"commit"`
	XXX_unrecognized []byte `json:"-"`
}
```
InitialState方法直接返回hardstate和快照里元数据记录的confstate实例
```go
func (ms *MemoryStorage) InitialState() (pb.HardState, pb.ConfState, error) {
	return ms.hardState, ms.snapshot.Metadata.ConfState, nil
}
``` 
LastIndex和FirstIndex分别返回ents数组的最后和第一条
```
// LastIndex implements the Storage interface.
func (ms *MemoryStorage) LastIndex() (uint64, error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.lastIndex(), nil
}

func (ms *MemoryStorage) lastIndex() uint64 {
	return ms.ents[0].Index + uint64(len(ms.ents)) - 1
}

// FirstIndex implements the Storage interface.
func (ms *MemoryStorage) FirstIndex() (uint64, error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.firstIndex(), nil
}

func (ms *MemoryStorage) firstIndex() uint64 {
	return ms.ents[0].Index + 1
}
```
ApplySnapshot  
更新快照数据，将snapshot实例保存到memorystorage中
```go
func (ms *MemoryStorage) ApplySnapshot(snap pb.Snapshot) error {
	ms.Lock()
	defer ms.Unlock()

	//判断index 新旧谁大，新的大才替换
	msIndex := ms.snapshot.Metadata.Index
	snapIndex := snap.Metadata.Index
	if msIndex >= snapIndex {
		return ErrSnapOutOfDate
	}

	ms.snapshot = snap
	ms.ents = []pb.Entry{{Term: snap.Metadata.Term, Index: snap.Metadata.Index}}
	return nil
}
```
Append  
向快照添加数据
```go
func (ms *MemoryStorage) Append(entries []pb.Entry) error {
    //常规检查
	if len(entries) == 0 {
		return nil
	}
	ms.Lock()
	defer ms.Unlock()
	first := ms.firstIndex()
	//获取待添加的最后一条的index值
	last := entries[0].Index + uint64(len(entries)) - 1
	// shortcut if there is no new entry.
	if last < first {
		return nil
	}
	// truncate compacted entries
	//截断已经记入snapshot中的
	if first > entries[0].Index {
		entries = entries[first-entries[0].Index:]
	}
    //计算切片一条和first的差
	offset := entries[0].Index - ms.ents[0].Index
	switch {
	case uint64(len(ms.ents)) > offset:
		ms.ents = append([]pb.Entry{}, ms.ents[:offset]...)
		ms.ents = append(ms.ents, entries...)
	case uint64(len(ms.ents)) == offset:
		ms.ents = append(ms.ents, entries...)
	default:
		raftLogger.Panicf("missing log entry [last: %d, append at: %d]",
			ms.lastIndex(), entries[0].Index)
	}
	return nil
}

//ps : 有点绕，为了不重复append
```
entries  
查询方法，查询指定范围的entry
```go
// Entries implements the Storage interface.
func (ms *MemoryStorage) Entries(lo, hi, maxSize uint64) ([]pb.Entry, error) {
	ms.Lock()
	defer ms.Unlock()
	offset := ms.ents[0].Index
	//常规判断
	if lo <= offset {
		return nil, ErrCompacted
	}
	if hi > ms.lastIndex()+1 {
		raftLogger.Panicf("entries' hi(%d) is out of bound lastindex(%d)", hi, ms.lastIndex())
	}
	// only contains dummy entries.
	if len(ms.ents) == 1 {
		return nil, ErrUnavailable
	}
    //获取下标的数据
	ents := ms.ents[lo-offset : hi-offset]
	//limitsize把超过大小的数据剔除
	return limitSize(ents, maxSize), nil
}

//limitsize把超过大小的数据剔除
func limitSize(ents []pb.Entry, maxSize uint64) []pb.Entry {
	if len(ents) == 0 {
		return ents
	}
	size := ents[0].Size()
	var limit int
	for limit = 1; limit < len(ents); limit++ {
		size += ents[limit].Size()
		if uint64(size) > maxSize {
			break
		}
	}
	return ents[:limit]
}
```

term
返回指定的index对应的ents值
```go
// Term implements the Storage interface.
func (ms *MemoryStorage) Term(i uint64) (uint64, error) {
	ms.Lock()
	defer ms.Unlock()
	offset := ms.ents[0].Index
	if i < offset {
		return 0, ErrCompacted
	}
	if int(i-offset) >= len(ms.ents) {
		return 0, ErrUnavailable
	}
	return ms.ents[i-offset].Term, nil
}
```

ents压缩  CreateSnapshot--Compact  
CreateSnapshot（更新快照信息）  
```go
func (ms *MemoryStorage) CreateSnapshot(i uint64, cs *pb.ConfState, data []byte) (pb.Snapshot, error) {
	ms.Lock()
	defer ms.Unlock()
	//常规检查，要比当前snapshot的index小，抛出异常
	if i <= ms.snapshot.Metadata.Index {
		return pb.Snapshot{}, ErrSnapOutOfDate
	}

	offset := ms.ents[0].Index
	//比ms.ent的last大，抛出异常
	if i > ms.lastIndex() {
		raftLogger.Panicf("snapshot %d is out of bound lastindex(%d)", i, ms.lastIndex())
	}

	ms.snapshot.Metadata.Index = i
	ms.snapshot.Metadata.Term = ms.ents[i-offset].Term
	if cs != nil {
		ms.snapshot.Metadata.ConfState = *cs
	}
	ms.snapshot.Data = data
	return ms.snapshot, nil
}
```
Compact（删除ents中指定索引之前的数据）
```go
func (ms *MemoryStorage) Compact(compactIndex uint64) error {
	ms.Lock()
	defer ms.Unlock()
	offset := ms.ents[0].Index
	//常规检查
	if compactIndex <= offset {
		return ErrCompacted
	}
	//常规检查
	if compactIndex > ms.lastIndex() {
		raftLogger.Panicf("compact %d is out of bound lastindex(%d)", compactIndex, ms.lastIndex())
	}
    //切片操作咯
	i := compactIndex - offset
	ents := make([]pb.Entry, 1, 1+uint64(len(ms.ents))-i)
	ents[0].Index = ms.ents[i].Index
	ents[0].Term = ms.ents[i].Term
	ents = append(ents, ms.ents[i+1:]...)
	ms.ents = ents
	return nil
}
```