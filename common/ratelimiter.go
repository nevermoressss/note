package common

import (
	"sync"
	"time"
)

type Limit float64

const InfDuration = time.Duration(1<<63 - 1)

type Limiter struct {
	limit  Limit
	cap    int
	mu     sync.Mutex
	tokens float64
	// 限制器的令牌字段的最后一次更新
	last time.Time
}

func NewLimiter(r Limit, cap int) *Limiter {
	return &Limiter{
		limit: r,
		cap:   cap,
	}
}

func (lim *Limiter) Wait() {
	now := time.Now()
	// Reserve
	r := lim.reserve(now)
	delay := r.DelayFrom(now)
	if delay < 0 {
		return
	}
	// Wait
	time.Sleep(r.DelayFrom(now))
}

func (lim *Limiter) Delay() time.Duration {
	now := time.Now()
	// Reserve
	r := lim.reserve(now)
	delay := r.DelayFrom(now)
	if delay < 0 {
		return 0
	}
	return r.DelayFrom(now)
}

func (lim *Limiter) reserve(now time.Time) Reservation {
	lim.mu.Lock()
	now, tokens := lim.advance(now)
	tokens -= 1
	// Calculate the wait duration
	var waitDuration time.Duration
	if tokens < 0 {
		waitDuration = lim.limit.durationFromTokens(-tokens)
	}

	// Prepare reservation
	r := Reservation{
		lim:   lim,
		limit: lim.limit,
	}
	r.timeToAct = now.Add(waitDuration)
	// Update state
	lim.last = now
	lim.tokens = tokens
	lim.mu.Unlock()
	return r
}

func (lim *Limiter) advance(now time.Time) (newNow time.Time, newTokens float64) {
	last := lim.last
	if now.Before(last) {
		last = now
	}
	// Avoid making delta overflow below when last is very old.
	maxElapsed := lim.limit.durationFromTokens(float64(lim.cap) - lim.tokens)
	elapsed := now.Sub(last)
	if elapsed > maxElapsed {
		elapsed = maxElapsed
	}
	// Calculate the new number of tokens, due to time that passed.
	delta := lim.limit.tokensFromDuration(elapsed)
	tokens := lim.tokens + delta
	if reserve := float64(lim.cap); tokens > reserve {
		tokens = reserve
	}
	return now, tokens
}

// durationFromTokens is a unit conversion function from the number of tokens to the duration
// of time it takes to accumulate them at a rate of limit tokens per second.
func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	seconds := tokens / float64(limit)
	return time.Nanosecond * time.Duration(1e9*seconds)
}

// tokensFromDuration is a unit conversion function from a time duration to the number of tokens
// which could be accumulated during that duration at a rate of limit tokens per second.
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	return d.Seconds() * float64(limit)
}

type Reservation struct {
	ok        bool
	lim       *Limiter
	timeToAct time.Time
	// This is the Limit at reservation time, it can change later.
	limit Limit
}

// DelayFrom returns the duration for which the reservation holder must wait
// before taking the reserved action.  Zero duration means act immediately.
// InfDuration means the limiter cannot grant the tokens requested in this
// Reservation within the maximum wait time.
func (r *Reservation) DelayFrom(now time.Time) time.Duration {
	delay := r.timeToAct.Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}
