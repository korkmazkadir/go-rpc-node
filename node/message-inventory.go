package node

import (
	"time"
)

type inventoryItem struct {
	lastAccessedTime int64
	message          *Message
}

// MessageInventory implements a TTL based message inventory
type messageInventory struct {
	ttl            int
	availableItems map[string]*inventoryItem
	lastClearTime  int64
}

// NewMessageInventory creates and initiate a message inventory
func newMessageInventory(timeToLiveInSeconds int) *messageInventory {
	mf := new(messageInventory)
	mf.ttl = timeToLiveInSeconds
	mf.availableItems = make(map[string]*inventoryItem)
	mf.lastClearTime = time.Now().Unix()
	return mf
}

// Add adds a message to the message inventory
func (m *messageInventory) Add(message *Message) bool {

	hash := message.Hash()
	_, result := m.availableItems[hash]
	time := time.Now().Unix()

	if result == true {
		m.availableItems[hash].lastAccessedTime = time
	} else {
		m.availableItems[hash] = &inventoryItem{lastAccessedTime: time, message: message}
	}

	m.clear(time)

	return !result
}

func (m *messageInventory) Get(messageHash string) *Message {

	inventoryItem := m.availableItems[messageHash]
	if inventoryItem == nil {
		return nil
	}

	return inventoryItem.message
}

func (m *messageInventory) clear(currentTime int64) {

	if (currentTime - m.lastClearTime) < int64(m.ttl) {
		return
	}

	for key, item := range m.availableItems {
		if (currentTime - item.lastAccessedTime) > int64(m.ttl) {
			delete(m.availableItems, key)
		}
	}

	m.lastClearTime = currentTime
}

func (m *messageInventory) size() int {
	return len(m.availableItems)
}
