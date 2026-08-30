package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowitem"
)

func testPaths(t *testing.T) *path.Paths {
	t.Helper()

	root := t.TempDir()
	paths, err := path.New(root)
	if err != nil {
		t.Fatalf("path.New() error = %v", err)
	}
	return paths
}

func saveItems(t *testing.T, filename string, items ...*wowitem.Item) {
	t.Helper()

	p := persist.New[int64, wowitem.Item](filename)
	for _, item := range items {
		p.Set(item.ID(), *item)
	}
	if err := p.Save(); err != nil {
		t.Fatalf("save items: %v", err)
	}
}

func saveAppearances(t *testing.T, filename string, appearances ...int64) {
	t.Helper()

	p := persist.New[int64, bool](filename)
	for _, id := range appearances {
		p.Set(id, true)
	}
	if err := p.Save(); err != nil {
		t.Fatalf("save appearances: %v", err)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	f()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestCreateItemPersist(t *testing.T) {
	paths := testPaths(t)

	output := captureStdout(t, func() {
		if err := createItemPersist(paths); err != nil {
			t.Fatalf("createItemPersist() error = %v", err)
		}
	})

	if !strings.Contains(output, "Saved items persist") {
		t.Fatalf("output = %q, want save message", output)
	}

	p, err := wowitem.New(paths.Items + ".new")
	if err != nil {
		t.Fatalf("load created persistence: %v", err)
	}
	if got := p.Len(); got != 0 {
		t.Fatalf("created persistence length = %d, want 0", got)
	}
}

func TestRunCreate(t *testing.T) {
	paths := testPaths(t)

	if err := runCreate([]string{"item"}, paths); err != nil {
		t.Fatalf("runCreate(item) error = %v", err)
	}

	if _, err := os.Stat(paths.Items + ".new.gob"); err != nil {
		t.Fatalf("created persistence: %v", err)
	}
}

func TestRunCreateInvalidPersistence(t *testing.T) {
	paths := testPaths(t)

	if err := runCreate([]string{"bogus"}, paths); err == nil || !strings.Contains(err.Error(), "unknown persistence type") {
		t.Fatalf("runCreate(bogus) error = %v", err)
	}
}

func TestRunCreateRequiresType(t *testing.T) {
	paths := testPaths(t)

	if err := runCreate(nil, paths); err == nil || !strings.Contains(err.Error(), "must specify a persistence type") {
		t.Fatalf("runCreate(nil) error = %v", err)
	}
}

func TestDeleteItem(t *testing.T) {
	paths := testPaths(t)
	item := wowitem.NewItem(map[string]any{
		"id":           json.Number("123"),
		"name":         "Test Item",
		"level":        json.Number("10"),
		"is_stackable": false,
		"item_class":   map[string]any{"name": "Armor"},
	})
	saveItems(t, paths.Items, item)

	output := captureStdout(t, func() {
		if err := deleteItem(item.ID(), paths); err != nil {
			t.Fatalf("deleteItem() error = %v", err)
		}
	})

	if !strings.Contains(output, "Deleted itemID: 123") {
		t.Fatalf("output = %q", output)
	}

	p, err := wowitem.New(paths.Items)
	if err != nil {
		t.Fatalf("reload persistence: %v", err)
	}
	if _, ok := p.Persistence.Get(item.ID()); ok {
		t.Fatal("deleted item unexpectedly remained in persistence")
	}
}

func TestRunDeleteRequiresID(t *testing.T) {
	paths := testPaths(t)

	if err := runDelete(nil, paths); err == nil || !strings.Contains(err.Error(), "delete requires -id") {
		t.Fatalf("runDelete(nil) error = %v", err)
	}
}

func TestJSON(t *testing.T) {
	paths := testPaths(t)
	item := wowitem.NewItem(map[string]any{
		"id":           json.Number("123"),
		"name":         "Test Item",
		"level":        json.Number("10"),
		"is_stackable": false,
		"item_class":   map[string]any{"name": "Armor"},
	})
	saveItems(t, paths.Items, item)

	output := captureStdout(t, func() {
		if err := asJSON(item.ID(), paths); err != nil {
			t.Fatalf("json() error = %v", err)
		}
	})

	if !strings.Contains(output, `"id": 123`) {
		t.Fatalf("output = %q, want item ID", output)
	}
	if !strings.Contains(output, `"name": "Test Item"`) {
		t.Fatalf("output = %q, want item name", output)
	}
}

func TestRunJSONRequiresID(t *testing.T) {
	paths := testPaths(t)

	if err := runJSON(nil, paths); err == nil || !strings.Contains(err.Error(), "json requires -id") {
		t.Fatalf("runJSON(nil) error = %v", err)
	}
}

func TestRunQueryByNameAndSort(t *testing.T) {
	paths := testPaths(t)

	items := []*wowitem.Item{
		wowitem.NewItem(map[string]any{
			"id":             json.Number("200"),
			"name":           "Alpha Sword",
			"level":          json.Number("10"),
			"inventory_type": "ONE_HANDED",
			"is_stackable":   false,
			"item_class":     map[string]any{"name": "Weapon"},
			"preview_item":   map[string]any{"quality": map[string]any{"name": "Rare"}},
		}),
		wowitem.NewItem(map[string]any{
			"id":             json.Number("300"),
			"name":           "Beta Sword",
			"level":          json.Number("20"),
			"inventory_type": "ONE_HANDED",
			"is_stackable":   false,
			"item_class":     map[string]any{"name": "Weapon"},
			"preview_item":   map[string]any{"quality": map[string]any{"name": "Epic"}},
		}),
	}
	saveItems(t, paths.Items, items...)
	saveAppearances(t, paths.Appearances)

	output := captureStdout(t, func() {
		if err := runQuery([]string{"-name", "sword", "-sort", "id"}, paths); err != nil {
			t.Fatalf("runQuery() error = %v", err)
		}
	})

	if !strings.Contains(output, "Alpha Sword") || !strings.Contains(output, "Beta Sword") {
		t.Fatalf("wanted both items: \n%q\n", output)
	}
	if strings.Index(output, "Beta Sword") < strings.Index(output, "Alpha Sword") {
		t.Fatalf("output sorted incorrectly: \n%q\n", output)
	}
}

func TestRunQueryInvalidSort(t *testing.T) {
	paths := testPaths(t)
	saveItems(t, paths.Items)
	saveAppearances(t, paths.Appearances)

	if err := runQuery([]string{"-sort", "bogus"}, paths); err == nil || !strings.Contains(err.Error(), "sort order must be one of") {
		t.Fatalf("runQuery() error = %v", err)
	}
}
