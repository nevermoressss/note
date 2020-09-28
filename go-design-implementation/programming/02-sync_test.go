package programming

import (
	"sync"
	"testing"
)

func TestSync(t *testing.T) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	mutex.Unlock()
	rwMutex := sync.RWMutex{}
	rwMutex.Lock()
	rwMutex.Unlock()
	rwMutex.RLock()
	rwMutex.RUnlock()
	group := sync.WaitGroup{}
	group.Add(1)
	once := sync.Once{}
	once.Do(
		func() {
			t.Log("hello word")
		})
	cond := sync.Cond{L: mutex}
	cond.Wait()
}
