package persist

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	p := New[string, int]("cache")

	if p == nil {
		t.Fatal("New() returned nil")
	}
	if got, want := p.Path(), "cache.gob"; got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
	if got := p.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
	if p.Dirty() {
		t.Fatal("new persistence is dirty")
	}
}

func TestSetGetAndDirty(t *testing.T) {
	p := New[string, int]("test")

	p.Set("one", 1)

	got, ok := p.Get("one")
	if !ok || got != 1 {
		t.Fatalf("Get() = (%d, %v), want (1, true)", got, ok)
	}
	if !p.Dirty() {
		t.Fatal("Set() did not mark persistence dirty")
	}
}

func TestGetMissingReturnsZeroValue(t *testing.T) {
	p := New[string, int]("test")

	got, ok := p.Get("missing")
	if ok {
		t.Fatal("Get() reported missing key as present")
	}
	if got != 0 {
		t.Fatalf("Get() = %d, want zero value", got)
	}
}

func TestSetOverwritesWithoutChangingLength(t *testing.T) {
	p := New[string, int]("test")

	p.Set("x", 1)
	p.Set("x", 2)

	if got, _ := p.Get("x"); got != 2 {
		t.Fatalf("Get() = %d, want 2", got)
	}
	if got := p.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
}

func TestDelete(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 10)
	if !p.Dirty() {
		t.Fatal("Set() did not mark persistence dirty")
	}

	p.Delete("a")

	if _, ok := p.Get("a"); ok {
		t.Fatal("Delete() left key present")
	}
	if got := p.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
	if !p.Dirty() {
		t.Fatal("Delete() did not leave persistence dirty")
	}
}

func TestDeleteMissingDoesNotMarkDirty(t *testing.T) {
	p := New[string, int]("test")

	p.Delete("missing")

	if p.Dirty() {
		t.Fatal("Delete() marked unchanged persistence dirty")
	}
}

func TestKeys(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 1)
	p.Set("b", 2)
	p.Set("c", 3)

	keys := p.Keys()
	slices.Sort(keys)

	if want := []string{"a", "b", "c"}; !slices.Equal(keys, want) {
		t.Fatalf("Keys() = %v, want %v", keys, want)
	}
}

func TestKeysReturnsIndependentSlice(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 1)
	p.Set("b", 2)

	keys := p.Keys()
	keys[0] = "changed"
	keys = append(keys, "extra")

	if got := p.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
	if _, ok := p.Get("changed"); ok {
		t.Fatal("mutating Keys() result changed persistence")
	}
}

func TestValues(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 3)
	p.Set("b", 1)
	p.Set("c", 2)

	values := p.Values()
	slices.Sort(values)

	if want := []int{1, 2, 3}; !slices.Equal(values, want) {
		t.Fatalf("Values() = %v, want %v", values, want)
	}
}

func TestValuesReturnsIndependentSlice(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 1)
	p.Set("b", 2)

	values := p.Values()
	values[0] = 999

	if got, ok := p.Get("a"); ok && got == 999 {
		t.Fatal("mutating Values() result changed persistence")
	}
	if got := p.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
}

func TestSearchFound(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 10)
	p.Set("b", 20)

	k, v, ok := p.Search(func(v int) bool { return v == 20 })
	if !ok || k != "b" || v != 20 {
		t.Fatalf("Search() = (%q, %d, %v), want (b, 20, true)", k, v, ok)
	}
}

func TestSearchNotFoundReturnsZeroValues(t *testing.T) {
	p := New[string, int]("test")
	p.Set("a", 1)

	k, v, ok := p.Search(func(v int) bool { return v == 99 })
	if ok {
		t.Fatal("Search() reported an unexpected match")
	}
	if k != "" || v != 0 {
		t.Fatalf("Search() = (%q, %d, %v), want zero key/value and false", k, v, ok)
	}
}

func TestLoadReplacesDataAndClearsDirty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache")

	stored := New[string, int](path)
	stored.Set("disk", 10)
	if err := stored.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	p := New[string, int](path)
	p.Set("old", 99)
	if !p.Dirty() {
		t.Fatal("Set() did not mark persistence dirty")
	}

	if err := p.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, ok := p.Get("old"); ok {
		t.Fatal("Load() did not replace old data")
	}
	if got, ok := p.Get("disk"); !ok || got != 10 {
		t.Fatalf("Get(disk) = (%d, %v), want (10, true)", got, ok)
	}
	if p.Dirty() {
		t.Fatal("successful Load() left persistence dirty")
	}
}

func TestLoadMissingFile(t *testing.T) {
	p := New[string, int](filepath.Join(t.TempDir(), "missing"))

	if err := p.Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadErrorLeavesExistingDataUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad")
	p := New[string, int](path)
	p.Set("keep", 42)

	if err := os.WriteFile(p.Path(), []byte("not a gob"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := p.Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if got, ok := p.Get("keep"); !ok || got != 42 {
		t.Fatalf("existing data after failed Load() = (%d, %v), want (42, true)", got, ok)
	}
	if !p.Dirty() {
		t.Fatal("failed Load() incorrectly cleared dirty state")
	}
}

func TestLoadCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.gob")
	if err := os.WriteFile(path, []byte("not a gob"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := New[string, int](strings.TrimSuffix(path, ".gob"))
	if err := p.Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache")

	want := map[string]int{"one": 1, "two": 2, "three": 3}
	p := New[string, int](path)
	for k, v := range want {
		p.Set(k, v)
	}

	if err := p.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if p.Dirty() {
		t.Fatal("successful Save() left persistence dirty")
	}
	if _, err := os.Stat(p.Path()); err != nil {
		t.Fatalf("saved file does not exist: %v", err)
	}
	if _, err := os.Stat(p.Path() + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary file still exists: %v", err)
	}

	loaded := New[string, int](path)
	if err := loaded.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	for k, wantValue := range want {
		if got, ok := loaded.Get(k); !ok || got != wantValue {
			t.Fatalf("Get(%q) = (%d, %v), want (%d, true)", k, got, ok, wantValue)
		}
	}
}

func TestSaveEmptyPersistence(t *testing.T) {
	p := New[string, int](filepath.Join(t.TempDir(), "empty"))

	if err := p.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded := New[string, int](strings.TrimSuffix(p.Path(), ".gob"))
	if err := loaded.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := loaded.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
	if loaded.Dirty() {
		t.Fatal("loaded empty persistence is dirty")
	}
}

func TestSaveEncodingFailurePreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache")

	good := New[string, int](path)
	good.Set("value", 123)
	if err := good.Save(); err != nil {
		t.Fatalf("initial Save() error = %v", err)
	}
	original, err := os.ReadFile(good.Path())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	bad := New[string, any](path)
	bad.Set("bad", func() {})
	if err := bad.Save(); err == nil {
		t.Fatal("Save() error = nil, want encoding error")
	}
	if !bad.Dirty() {
		t.Fatal("failed Save() cleared dirty state")
	}

	current, err := os.ReadFile(good.Path())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !slices.Equal(current, original) {
		t.Fatal("failed Save() modified the existing persistence file")
	}
	if _, err := os.Stat(bad.Path() + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary file remains after failed Save(): %v", err)
	}
}

func TestGobRegistrationSupportsAnyMaps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "any")
	p := New[string, any](path)
	p.Set("map", map[string]any{"answer": 42})

	if err := p.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded := New[string, any](path)
	if err := loaded.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got, ok := loaded.Get("map")
	if !ok {
		t.Fatal("map entry missing after Load()")
	}
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("loaded value has type %T, want map[string]any", got)
	}
	if m["answer"] != 42 {
		t.Fatalf("loaded map = %v, want answer=42", m)
	}
}

func TestConcurrentAccess(t *testing.T) {
	p := New[int, int](filepath.Join(t.TempDir(), "concurrent"))

	const writers = 8
	const readers = 8
	const iterations = 500

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for g := range writers {
		go func(g int) {
			defer wg.Done()
			for i := range iterations {
				key := (g * iterations) + i
				p.Set(key, key)
			}
		}(g)
	}

	for g := range readers {
		go func(g int) {
			defer wg.Done()
			for i := range iterations {
				key := (g * iterations) + i
				_, _ = p.Get(key)
				_ = p.Len()
				_ = p.Dirty()
			}
		}(g)
	}

	wg.Wait()

	if got, want := p.Len(), writers*iterations; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
}
