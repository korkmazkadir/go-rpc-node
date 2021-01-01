package filter

import (
	"sync"
	"time"
)

// UniqueMessageFilter implements a TTL based filter
type UniqueMessageFilter struct {
	ttl               int
	processedMessages map[string]int64
	lastClearTime     int64
	mutex             sync.Mutex
}

// NewUniqueMessageFilter creates and initiate a message filter
func NewUniqueMessageFilter(timeToLiveInSeconds int) *UniqueMessageFilter {
	mf := new(UniqueMessageFilter)
	mf.ttl = timeToLiveInSeconds
	mf.processedMessages = make(map[string]int64)
	mf.lastClearTime = time.Now().Unix()
	mf.mutex = sync.Mutex{}
	return mf
}

// IfNotContainsAdd try to adds a message to filter if it is not available already
func (u *UniqueMessageFilter) IfNotContainsAdd(hash string) bool {

	u.mutex.Lock()
	defer u.mutex.Unlock()

	_, result := u.processedMessages[hash]

	time := time.Now().Unix()
	u.processedMessages[hash] = time
	u.clear(time)

	return !result
}

func (u *UniqueMessageFilter) clear(currentTime int64) {

	if (currentTime - u.lastClearTime) < int64(u.ttl) {
		return
	}

	for key, value := range u.processedMessages {
		if (currentTime - value) > int64(u.ttl) {
			delete(u.processedMessages, key)
		}
	}

	u.lastClearTime = currentTime
}

func (u *UniqueMessageFilter) size() int {
	return len(u.processedMessages)
}
