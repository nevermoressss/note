#### BigCache
```go
type BigCache struct {
	shards       []*cacheShard 	//分片
	lifeWindow   uint64			//过期时间
	clock        clock			//时间获取的方法
	hash         Hasher			//hash算法
	config       Config			//配置文件
	shardMask    uint64			//常量：分片数-1，用于&运算替换取模运算
	maxShardSize uint32			//分片最大容量
	close        chan struct{}	//控制关闭
}
```
cacheShard：
```go
type BigCache struct {
	hashmap     map[uint64]uint32		//uint64:key的hash值，uint32：存储在bytesQueue的开始下标
	entries     queue.BytesQueue		//实际存储的结构,字节数组，一个环形队列
	lock        sync.RWMutex			//读写锁
	entryBuffer []byte					//用于set的时候临时存放要set的value值
	onRemove    onRemoveCallback		//触发删除时的回调函数
	isVerbose    bool					//是否打日志
	statsEnabled bool					//是否计算请求缓存资源的次数
	logger       Logger					//日志
	clock        clock					//时间获取的方法
	lifeWindow   uint64					//过期时间
	hashmapStats map[uint64]uint32		//计数
	stats        Stats					//计数的结构
}
```
BytesQueue：
```go
type BytesQueue struct {
	full            bool		//是否满了
	array           []byte		//存储的地方byte数组
	capacity        int			//当前容量
	maxCapacity     int			//最大可扩容到多少，<=0无限扩容
	head            int			//head下标
	tail            int			//tail下标
	count           int			//
	rightMargin     int
	headerBuffer    []byte		//约定：每个value存储时要加header信息
	verbose         bool		//日志
	initialCapacity int
}
```
分析下get和set这2个方法吧
set：
```go
// Set saves entry under the key
func (c *BigCache) Set(key string, entry []byte) error {
	hashedKey := c.hash.Sum64(key)				//计算key的hash值
	shard := c.getShard(hashedKey)				//确定分片
	return shard.set(key, hashedKey, entry) 	//执行分片的set方法
}
```
计算hash值没什么好说的，确定分片这里用了一个巧妙的位运算来代替了模运算

hashedKey&c.shardMask  你品   			    8%4 和 8&3 

shard.set：
```go
func (s *cacheShard) set(key string, hashedKey uint64, entry []byte) error {
	currentTimestamp := uint64(s.clock.epoch()) //先取个当前时间
	s.lock.Lock()//该分片加锁
	//看下这个hash值存不存在，hashmap的key是hash值，value对应的是queue的下标。
	//ps：一开始分析这里的时候，发现是通过判断下标是否等于0来判断存不存在的。当时我想，第一个值的下标不就是0吗，那第一个值岂不是判断不能判断是否存在？？.........后来看到bytes_queue.go有一个常量，leftMarginIndex = 1，索引从1开始，不使用0下标......兄弟，套路有点多啊.
	if previousIndex := s.hashmap[hashedKey]; previousIndex != 0 {
		if previousEntry, err := s.entries.Get(int(previousIndex)); err == nil {
		//遇到hash冲突了，先清除旧的值
			resetKeyFromEntry(previousEntry)
		}
	}
	//取队列里的第一个值，执行onEvict方法，检查第一个key是否过期，过期就执行删除的回调方法，删除
	//ps:假设同时符合hash冲突和第一个元素并且过期，那边回调函数是不是就拿不到过期key的信息了？因为在上面hash冲突的时候已经把旧的值清掉了，很边缘的情况。有没有有志之士验证一下哈
	if oldestEntry, err := s.entries.Peek(); err == nil {
		s.onEvict(oldestEntry, currentTimestamp, s.removeOldestEntry)
	}
	//构造实际存储的byte数组
	w := wrapEntry(currentTimestamp, hashedKey, key, entry, &s.entryBuffer)
	//
	for {
		if index, err := s.entries.Push(w); err == nil {
			//插入成功
			s.hashmap[hashedKey] = uint32(index)
			s.lock.Unlock()
			return nil
		}
		//插入不成功（队列full了）就淘汰队头元素，for继续插入
		if s.removeOldestEntry(NoSpace) != nil {
			s.lock.Unlock()
			return fmt.Errorf("entry is bigger than max shard size")
		}
	}
}
```
get：
```go
func (c *BigCache) Get(key string)  error {
	hashedKey := c.hash.Sum64(key)				//计算key的hash值
	shard := c.getShard(hashedKey)				//确定分片
	return shard.get(key, hashedKey)	 		//执行分片的get方法
}
```
shard.get：
```go
func (s *cacheShard) get(key string, hashedKey uint64) ([]byte, error) {
	s.lock.RLock()			//读锁
	//getWrappedEntry，通过hashedkey获取array下标，并且把对应的数据读取出来
	wrappedEntry, err := s.getWrappedEntry(hashedKey)
	if err != nil {
		s.lock.RUnlock()
		return nil, err
	}
	//解析信息判断存储的key和传进来的key是否是同一个
	if entryKey := readKeyFromEntry(wrappedEntry); key != entryKey {
		//不相同，value的值不是当前key对应的值，而是与其hash冲突的另外一个key的值
		s.lock.RUnlock()
		s.collision()
		if s.isVerbose {
			s.logger.Printf("Collision detected. Both %q and %q have the same hash %x", key, entryKey, hashedKey)
		}
		return nil, ErrEntryNotFound
	}
	//相同就解析出value值返回
	entry := readEntry(wrappedEntry)
	s.lock.RUnlock()
	s.hit(hashedKey)

	return entry, nil
}
```