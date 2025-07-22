package evictors

import (
	"container/list"
	"sync"
)

type entry struct {
	key string
}

type LRUEvictor struct {
	mu    sync.Mutex
	items map[string]*list.Element
	list  *list.List
}

// NewLRUEvictor creates a new LRU eviction policy instance.
func NewLRUEvictor() *LRUEvictor {
	return &LRUEvictor{
		items: make(map[string]*list.Element),
		list:  list.New(),
	}
}

func (l *LRUEvictor) OnSet(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if el, ok := l.items[key]; ok {
		l.list.MoveToFront(el)
		return
	}

	l.items[key] = l.list.PushFront(&entry{key: key})
}

func (l *LRUEvictor) OnGet(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if el, ok := l.items[key]; ok {
		l.list.MoveToFront(el)
	}
}

func (l *LRUEvictor) OnDelete(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if el, ok := l.items[key]; ok {
		l.list.Remove(el)
		delete(l.items, key)
	}
}

// Evict returns the n least recently used keys, 
// where n is passed as count.
// The cache will then delete these keys in batch.
// If no keys are available to evict, it returns nil.
func (l *LRUEvictor) Evict(count int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.list.Len() == 0 {
		return nil
	}

	keysToEvict := make([]string, 0, count)
	for i := 0; i < count; i++ {
		back := l.list.Back()
		if back == nil {
			break
		}

		l.list.Remove(back)
		e := back.Value.(*entry)
		delete(l.items, e.key)

		keysToEvict = append(keysToEvict, e.key)
	}

	return keysToEvict
}
