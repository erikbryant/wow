package wowitem

import "testing"

func TestPersistenceSearchAndSortedKeys(t *testing.T) {
	p := NewEmpty(t.TempDir() + "/items")
	p.Set(20, NewItem(map[string]any{"id": float64(20), "name": "Beta"}))
	p.Set(10, NewItem(map[string]any{"id": float64(10), "name": "Alpha"}))
	if got := p.Keys(); len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Fatalf("keys=%v", got)
	}
	if got := p.Search("Beta"); got.ID() != 20 {
		t.Fatalf("search=%d", got.ID())
	}
	if got := p.Search("Missing"); got.ID() != 0 {
		t.Fatalf("missing search=%d", got.ID())
	}
}
