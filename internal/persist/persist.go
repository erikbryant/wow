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
}

func New[K comparable, V any](name string) *Persistence[K, V] {
	gob.Register(map[string]any{})
	gob.Register([]any{})

	return &Persistence[K, V]{
		filename: filepath.Join(dataDirectory, name+".gob"),
		data:     nil,
	}
}

// requireLoaded panics if the persistence has not been loaded; caller responsible for acquiring mu.lock
func (c *Persistence[K, V]) requireLoaded() {
	if c.data == nil {
		panic("persist: persistence has not been loaded: " + c.filename)
	}
}

func (c *Persistence[K, V]) Load() error {
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

func (c *Persistence[K, V]) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.requireLoaded()

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

func (c *Persistence[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.requireLoaded()
	return len(c.data)
}

func (c *Persistence[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.requireLoaded()
	v, ok := c.data[key]
	return v, ok
}

func (c *Persistence[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requireLoaded()
	c.data[key] = value
}

func (c *Persistence[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requireLoaded()
	delete(c.data, key)
}

func (c *Persistence[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.requireLoaded()

	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}

	return keys
}

func (c *Persistence[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.requireLoaded()

	values := make([]V, 0, len(c.data))
	for _, v := range c.data {
		values = append(values, v)
	}

	return values
}

func (c *Persistence[K, V]) Search(searchFunc func(v V) bool) (K, V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.requireLoaded()

	for k, v := range c.data {
		if searchFunc(v) {
			return k, v, true
		}
	}

	var zeroV V
	var zeroK K
	return zeroK, zeroV, false
}
