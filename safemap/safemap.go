package safemap

import (
	"sync"
)

// SafeMap sm
type SafeMap struct {
	rw *sync.RWMutex
	c  map[interface{}]interface{}
}

// New create a new sm
func New() *SafeMap {
	return &SafeMap{
		rw: &sync.RWMutex{},
		c:  make(map[interface{}]interface{}),
	}
}

// Set set k,v
func (sm *SafeMap) Set(key, value interface{}) {
	sm.rw.Lock()
	sm.c[key] = value
	sm.rw.Unlock()
}

// Get v using k
func (sm *SafeMap) Get(key interface{}) (interface{}, bool) {
	sm.rw.RLock()
	defer sm.rw.RUnlock()
	v, ok := sm.c[key]
	return v, ok
}

// Exist check if key exist
func (sm *SafeMap) Exist(key interface{}) bool {
	sm.rw.RLock()
	defer sm.rw.RUnlock()
	_, ok := sm.c[key]
	return ok
}

// Size sm size
func (sm *SafeMap) Size() int {
	sm.rw.RLock()
	defer sm.rw.RUnlock()
	return len(sm.c)
}

// Delete remove key from sm
func (sm *SafeMap) Delete(key interface{}) {
	sm.rw.Lock()
	delete(sm.c, key)
	sm.rw.Unlock()
}

// Items all items of sm
func (sm *SafeMap) Items() map[interface{}]interface{} {
	sm.rw.RLock()
	defer sm.rw.RUnlock()
	// copy
	m := make(map[interface{}]interface{}, len(sm.c))
	for k, v := range sm.c {
		m[k] = v
	}
	return m
}
