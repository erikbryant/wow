package persist

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
)

type Persistence[K comparable, V any] struct {
	filename string
	mu       sync.RWMutex
	data     map[K]V
	dirty    bool
}

func init() {
	// All stored data is in this format.
	gob.Register(map[string]any{})

	// Some stored data contains lists.
	gob.Register([]any{})
}

// New creates a new Persistence backed by persistencePath + ".gob".
func New[K comparable, V any](persistencePath string) *Persistence[K, V] {
	return &Persistence[K, V]{
		filename: persistencePath + ".gob",
		data:     make(map[K]V),
		dirty:    false,
	}
}

// Load replaces the current data with the contents of the persistence file.
func (p *Persistence[K, V]) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.Open(p.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)

	var data map[K]V
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	if data == nil {
		data = make(map[K]V)
		fmt.Fprintf(os.Stderr, "persistence data loaded, but is empty: %v", p.filename)
	}

	p.data = data
	p.dirty = false

	return nil
}

// Dirty reports whether the persistence has been modified since the last
// successful Load or Save.
func (p *Persistence[K, V]) Dirty() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dirty
}

// Save writes the current data to disk atomically.
func (p *Persistence[K, V]) Save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	tmp := p.filename + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(f)

	if err := encoder.Encode(p.data); err != nil {
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

	err = os.Rename(tmp, p.filename)
	if err != nil {
		os.Remove(tmp)
		return err
	}

	// If we got here then we have a clean save! :)
	p.dirty = false

	return nil
}

// Len returns the number of entries.
func (p *Persistence[K, V]) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.data)
}

// Get returns the value associated with key.
//
// Values returned by Persistence must be treated as immutable.
func (p *Persistence[K, V]) Get(key K) (V, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.data[key]

	return v, ok
}

// Set associates value with key and marks the persistence dirty.
func (p *Persistence[K, V]) Set(key K, value V) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[key] = value
	p.dirty = true
}

// Delete removes key, if present, and marks the persistence dirty.
func (p *Persistence[K, V]) Delete(key K) {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.data[key]
	if ok {
		delete(p.data, key)
		p.dirty = true
	}
}

// Keys returns all keys.
func (p *Persistence[K, V]) Keys() []K {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]K, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}

	return keys
}

// Values returns all values.
//
// Values returned by Persistence must be treated as immutable.
func (p *Persistence[K, V]) Values() []V {
	p.mu.RLock()
	defer p.mu.RUnlock()

	values := make([]V, 0, len(p.data))
	for _, v := range p.data {
		values = append(values, v)
	}

	return values
}

// Search returns the first entry for which searchFunc returns true.
func (p *Persistence[K, V]) Search(searchFunc func(v V) bool) (K, V, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for k, v := range p.data {
		if searchFunc(v) {
			return k, v, true
		}
	}

	var zeroV V
	var zeroK K
	return zeroK, zeroV, false
}

// Path returns the path to the backing persistence file.
func (p *Persistence[K, V]) Path() string {
	return p.filename
}
