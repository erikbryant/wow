package persist

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
)

const (
	dataDirectory = "./data"
)

type Persistence[K comparable, V any] struct {
	filename string
	mu       sync.RWMutex
	data     map[K]V
	loaded   bool
	dirty    bool
}

func New[K comparable, V any](name string) *Persistence[K, V] {
	gob.Register(map[string]any{})
	gob.Register([]any{})

	return &Persistence[K, V]{
		filename: filepath.Join(dataDirectory, name+".gob"),
		data:     make(map[K]V),
		loaded:   false,
		dirty:    false,
	}
}

func (c *Persistence[K, V]) Loaded() bool {
	return c.loaded
}

func (c *Persistence[K, V]) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Loaded() {
		return nil
	}

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
	c.loaded = true
	c.dirty = false

	return nil
}

func (c *Persistence[K, V]) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		// Nothing changed, no need to save
		return nil
	}

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

	err = f.Sync()
	if err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	err = f.Close()
	if err != nil {
		os.Remove(tmp)
		return err
	}

	err = os.Rename(tmp, c.filename)
	if err != nil {
		os.Remove(tmp)
		return err
	}

	// If we got here then we have a clean save! :)
	c.dirty = false

	return nil
}

func (c *Persistence[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *Persistence[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]

	// Values returned by Persistence must be treated as immutable.
	return v, ok
}

func (c *Persistence[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	c.dirty = true
}

func (c *Persistence[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.data[key]
	if ok {
		delete(c.data, key)
		c.dirty = true
	}
}

func (c *Persistence[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}

	return keys
}

func (c *Persistence[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.data))
	for _, v := range c.data {
		values = append(values, v)
	}

	return values
}

func (c *Persistence[K, V]) Search(searchFunc func(v V) bool) (K, V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for k, v := range c.data {
		if searchFunc(v) {
			return k, v, true
		}
	}

	var zeroV V
	var zeroK K
	return zeroK, zeroV, false
}
