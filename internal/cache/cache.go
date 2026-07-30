package cache

import (
	"encoding/gob"
	"os"
	"sync"
)

type Cache[K comparable, V comparable] struct {
	filename string
	mu       sync.RWMutex
	data     map[K]V
}

func New[K comparable, V comparable](filename string) *Cache[K, V] {
	gob.Register(map[string]any{})
	gob.Register([]any{})

	return &Cache[K, V]{
		filename: filename,
		data:     make(map[K]V),
	}
}

func (c *Cache[K, V]) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.Open(c.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)

	var data map[K]V
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	c.data = data

	return nil
}

func (c *Cache[K, V]) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tmp := c.filename + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(f)

	if err := encoder.Encode(c.data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, c.filename)
}

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *Cache[K, V]) ReverseLookup(value V) (K, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for k, v := range c.data {
		if v == value {
			return k, true
		}
	}

	var zero K
	return zero, false
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
