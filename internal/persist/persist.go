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

	var data map[K]V

	if err := gob.NewDecoder(f).Decode(&data); err != nil {
		return err
	}

	if data == nil {
		data = make(map[K]V)
		fmt.Fprintf(os.Stderr, "persistence data loaded, but is empty: %s\n", p.filename)
	}

	p.data = data
	p.dirty = false

	return nil
}

// Dirty reports whether the persistence has been modified since the last
// successful Load or Save.
func (p *Persistence[K, V]) Dirty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

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

	if err := gob.NewEncoder(f).Encode(p.data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, p.filename); err != nil {
		_ = os.Remove(tmp)
		return err
	}

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

	value, ok := p.data[key]
	return value, ok
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

	if _, ok := p.data[key]; !ok {
		return
	}

	delete(p.data, key)
	p.dirty = true
}

// Keys returns all keys.
func (p *Persistence[K, V]) Keys() []K {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]K, 0, len(p.data))

	for key := range p.data {
		keys = append(keys, key)
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

	for _, value := range p.data {
		values = append(values, value)
	}

	return values
}

// Search returns the first entry for which searchFunc returns true.
func (p *Persistence[K, V]) Search(searchFunc func(V) bool) (K, V, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for key, value := range p.data {
		if searchFunc(value) {
			return key, value, true
		}
	}

	var zeroK K
	var zeroV V

	return zeroK, zeroV, false
}

// Path returns the path to the backing persistence file.
func (p *Persistence[K, V]) Path() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.filename
}
