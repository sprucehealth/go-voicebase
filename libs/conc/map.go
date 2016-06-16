package conc

import "sync"

// Map is a concurrent implementation of a map
type Map struct {
	mux  sync.RWMutex
	cmap map[string]interface{}
}

// NewMap returns a map for concurrent access.
func NewMap() *Map {
	return &Map{
		cmap: make(map[string]interface{}),
	}
}

// Get returns a value from the map or nil if the key does not exist
func (c *Map) Get(key string) interface{} {
	if c == nil {
		return nil
	}
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.cmap[key]
}

// Set sets, possibly overwriting, a value in the map
func (c *Map) Set(key string, value interface{}) {
	if c == nil {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.cmap[key] = value
}

// Delete deletes a value from the map
func (c *Map) Delete(key string) {
	if c == nil {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.cmap, key)
}

// Transact locks the map, calls the provided functions, and unlocks the map on return.
func (c *Map) Transact(fn func(map[string]interface{})) {
	if c == nil {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	fn(c.cmap)
}

// Clear locks the map and deletes all entries
func (c *Map) Clear() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.cmap = make(map[string]interface{})
}

// Snapshot returns a copy of the underlying values
func (c *Map) Snapshot() map[string]interface{} {
	if c == nil {
		return nil
	}
	c.mux.RLock()
	defer c.mux.RUnlock()
	snapshot := make(map[string]interface{}, len(c.cmap))
	for k, v := range c.cmap {
		snapshot[k] = v
	}
	return snapshot
}
