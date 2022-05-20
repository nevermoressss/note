package common

type limiter interface {
	// 等待获取令牌
	Wait()
	// 非阻塞尝试获取
	NotWait() bool
	// 归还
	Release()
	// 关闭
	Close()
}

// channal实现限流
type channelLimiter struct {
	ch chan struct{}
}

// 初始化
func NewChannelLimiter(limiter int) limiter {
	return &channelLimiter{
		make(chan struct{}, limiter),
	}
}

// 阻塞等待令牌
func (cl *channelLimiter) Wait() {
	cl.ch <- struct{}{}
}

// 非阻塞 通过返回值判断是否拿到令牌
func (cl *channelLimiter) NotWait() bool {
	select {
	case cl.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// 释放
func (cl *channelLimiter) Release() {
	<-cl.ch
}

// 关闭
func (cl *channelLimiter) Close() {
	close(cl.ch)
}
