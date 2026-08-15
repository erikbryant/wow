package persist

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	p := New[string, int]("cache")

	if p == nil {
		t.Fatal("New returned nil")
	}

	if got, want := p.filename, "cache"+".gob"; got != want {
		t.Fatalf("filename = %q, want %q", got, want)
	}

	if p.dirty {
		t.Fatal("new persistence should not be dirty")
	}

	if p.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", p.Len())
	}
}

func TestSetGet(t *testing.T) {
	p := New[string, int]("test")

	p.Set("one", 1)

	got, ok := p.Get("one")
	if !ok {
		t.Fatal("expected key to exist")
	}

	if got != 1 {
		t.Fatalf("Get() = %d, want 1", got)
	}

	if !p.dirty {
		t.Fatal("Set should mark persistence dirty")
	}
}

func TestGetMissing(t *testing.T) {
	p := New[string, int]("test")

	v, ok := p.Get("missing")

	if ok {
		t.Fatal("expected missing key")
	}

	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

func TestOverwrite(t *testing.T) {
	p := New[string, int]("test")

	p.Set("x", 1)
	p.Set("x", 2)

	if got, _ := p.Get("x"); got != 2 {
		t.Fatalf("got %d want 2", got)
	}

	if p.Len() != 1 {
		t.Fatalf("Len()=%d want 1", p.Len())
	}
}

func TestDelete(t *testing.T) {
	p := New[string, int]("test")

	p.Set("a", 10)
	p.Delete("a")

	if _, ok := p.Get("a"); ok {
		t.Fatal("key still exists")
	}

	if p.Len() != 0 {
		t.Fatal("expected empty persistence")
	}
}

func TestDeleteMissing(t *testing.T) {
	p := New[string, int]("test")

	p.Delete("missing")

	if p.dirty {
		t.Fatal("Delete marked unchanged data 'dirty'")
	}
}

func TestKeys(t *testing.T) {
	p := New[string, int]("test")

	p.Set("a", 1)
	p.Set("b", 2)
	p.Set("c", 3)

	keys := p.Keys()
	slices.Sort(keys)

	want := []string{"a", "b", "c"}

	if !slices.Equal(keys, want) {
		t.Fatalf("Keys()=%v want %v", keys, want)
	}
}

func TestValues(t *testing.T) {
	p := New[string, int]("test")

	p.Set("a", 3)
	p.Set("b", 1)
	p.Set("c", 2)

	values := p.Values()
	slices.Sort(values)

	want := []int{1, 2, 3}

	if !slices.Equal(values, want) {
		t.Fatalf("Values()=%v want %v", values, want)
	}
}

func TestSearchFound(t *testing.T) {
	p := New[string, int]("test")

	p.Set("a", 10)
	p.Set("b", 20)

	k, v, ok := p.Search(func(v int) bool {
		return v == 20
	})

	if !ok {
		t.Fatal("expected match")
	}

	if k != "b" || v != 20 {
		t.Fatalf("got (%q,%d)", k, v)
	}
}

func TestSearchNotFound(t *testing.T) {
	p := New[string, int]("test")

	p.Set("a", 1)

	k, v, ok := p.Search(func(v int) bool {
		return v == 99
	})

	if ok {
		t.Fatal("unexpected match")
	}

	if k != "" {
		t.Fatalf("expected zero key, got %q", k)
	}

	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

func TestLoadMissingFile(t *testing.T) {
	p := New[string, int]("missing")
	p.filename = filepath.Join(t.TempDir(), "missing.gob")

	if err := p.Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadCorruptFile(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "bad.gob")

	if err := os.WriteFile(file, []byte("not a gob"), 0644); err != nil {
		t.Fatal(err)
	}

	p := New[string, int]("bad")
	p.filename = file

	if err := p.Load(); err == nil {
		t.Fatal("expected decode error")
	}
}
