package query

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/erikbryant/wow/internal/wowitem"
)

func jsonNumber(n int64) json.Number {
	s := strconv.FormatInt(n, 10)
	return json.Number(s)
}

func qi(id int64, name, quality, class string, level int64) wowitem.Item {
	return wowitem.Item{XID: id, XItem: map[string]any{"id": jsonNumber(id), "name": name, "level": jsonNumber(level), "is_stackable": false, "item_class": map[string]any{"name": class}, "preview_item": map[string]any{"quality": map[string]any{"name": quality}}}}
}

func TestFind(t *testing.T) {
	items := []wowitem.Item{qi(1, "Alpha", "Rare", "Armor", 100), qi(2, "Beta", "Epic", "Weapon", 200), qi(3, "alpha sword", "Common", "Weapon", 150)}
	if got := Find(items); len(got) != 3 {
		t.Fatalf("no predicates: %d", len(got))
	}
	got := Find(items, NameContains("ALPHA"), ItemLevelAtLeast(100), ItemLevelAtMost(150))
	if len(got) != 2 || got[0].ID() != 1 || got[1].ID() != 3 {
		t.Fatalf("unexpected results: %v", got)
	}
	if len(Find(items, Rare(), Epic())) != 0 {
		t.Error("AND semantics expected")
	}
	if len(Find(items, ItemClass("Weapon"))) != 2 {
		t.Error("class predicate")
	}
	if len(Find(items, ItemID(2))) != 1 {
		t.Error("id predicate")
	}
}

func TestSort(t *testing.T) {
	items := []wowitem.Item{qi(3, "Charlie", "Common", "Armor", 1), qi(1, "Alpha", "Common", "Armor", 1), qi(2, "Beta", "Common", "Armor", 1)}
	Sort(items, ByID)
	if items[0].ID() != 1 || items[2].ID() != 3 {
		t.Fail()
	}
	Sort(items, ByName)
	if items[0].Name() != "Alpha" || items[2].Name() != "Charlie" {
		t.Fail()
	}
}
