package cache

import (
	"encoding/gob"
	"os"
	"sync"
)

// Callers use:
//
// cache.New(filename, map[string]any{})
// cache.Load()
//
// cache.Update(func(m map[int64]Item) {
//	m[id] = item
// })
//
// var item Item
// var ok bool
// cache.Get(func(m map[int64]Item) {
//	item, ok = m[id]
// })

type Cache[T any] struct {
	filename string
	mu       sync.RWMutex
	Data     T
}

func New[T any](filename string, initial T) *Cache[T] {
	gob.Register(map[string]any{})
	gob.Register([]any{})

	return &Cache[T]{
		filename: filename,
		Data:     initial,
	}
}

func (c *Cache[T]) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.Open(c.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)

	var data T
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	c.Data = data

	return nil
}

func (c *Cache[T]) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tmp := c.filename + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(f)

	if err := encoder.Encode(c.Data); err != nil {
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

func (c *Cache[T]) Get(fn func(T)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fn(c.Data)
}

func (c *Cache[T]) Update(fn func(T)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn(c.Data)
}
