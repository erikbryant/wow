package credentials

import (
	"errors"
	"testing"
)

type memoryStore struct {
	values map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		values: make(map[string]string),
	}
}

func (s *memoryStore) Get(name string) (string, error) {
	value, ok := s.values[name]
	if !ok {
		return "", ErrNotFound
	}

	return value, nil
}

func (s *memoryStore) Set(name, value string) error {
	if _, exists := s.values[name]; exists {
		return ErrExists
	}

	s.values[name] = value
	return nil
}

func (s *memoryStore) Delete(name string) error {
	if _, exists := s.values[name]; !exists {
		return ErrNotFound
	}

	delete(s.values, name)
	return nil
}

func TestStoreSetGet(t *testing.T) {
	store := newMemoryStore()

	err := store.Set("clientID", "test-client-id")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get("clientID")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "test-client-id" {
		t.Fatalf("Get() = %q, want %q", got, "test-client-id")
	}
}

func TestStoreDuplicate(t *testing.T) {
	store := newMemoryStore()

	if err := store.Set("clientID", "first"); err != nil {
		t.Fatalf("first Set() error = %v", err)
	}

	err := store.Set("clientID", "second")
	if !errors.Is(err, ErrExists) {
		t.Fatalf("second Set() error = %v, want ErrExists", err)
	}

	got, err := store.Get("clientID")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "first" {
		t.Fatalf("Get() = %q, want original value %q", got, "first")
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := newMemoryStore()

	_, err := store.Get("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestStoreDelete(t *testing.T) {
	store := newMemoryStore()

	if err := store.Set("clientSecret", "secret"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if err := store.Delete("clientSecret"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := store.Get("clientSecret")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestStoreDeleteMissing(t *testing.T) {
	store := newMemoryStore()

	err := store.Delete("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}
