package cache

import (
	"container/list"
	"sync"
)

// The cache uses the LRU algorithm and sets the maximum capacity for elimination.
type UniversalCache struct {
	// recently used queue
	recently *list.List
	// index to the queue node
	index    map[string]*list.Element
	nowBytes int64
	maxBytes int64
	mutex    sync.Mutex
}

// Value abstracts the interface of a cached value which could storage any type
// of data. And the cache will get the size by length of it byte slice.
type Value interface {
	Raw() *[]byte
}

type Item struct {
	key   string
	value Value
}

func NewUniversalCache(size int64) *UniversalCache {
	return &UniversalCache{
		recently: list.New(),
		index:    make(map[string]*list.Element),
		maxBytes: size,
	}
}

// This function doing as the usually LRU agorithm:
// If one item be hit, it will move to the front of queue.
// If nothing hits, nil and false will be returned.
func (u *UniversalCache) GET(key string) (interface{}, bool) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if node, hit := u.index[key]; hit {
		u.recently.MoveToFront(node)
		item := node.Value.(*Item)
		return item.value, true
	} else {
		return nil, false
	}
}

func (u *UniversalCache) SET(key string, value Value) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if node, hit := u.index[key]; hit {
		// if hit the cache to update
		item := node.Value.(*Item)
		u.nowBytes -= int64(len(*item.value.Raw())) + int64(len(*value.Raw()))
		item.value = value
		u.recently.MoveToFront(node)
	} else {
		// if not hit set the new one
		node := u.recently.PushFront(&Item{key, value})
		u.nowBytes += int64(len(key)) + int64(len(*value.Raw()))
		u.index[key] = node
	}
	// Eliminate redundant cache, when the nowBytes > maxBytes
	for u.nowBytes < u.maxBytes {
		node := u.recently.Back()
		if node != nil {
			item := node.Value.(*Item)
			delete(u.index, item.key)
			u.recently.Remove(node)
			u.nowBytes -= int64(len(item.key)) + int64(len(*item.value.Raw()))
		}
	}
}
