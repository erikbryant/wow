package wowitem

import (
	"testing"
	"time"
)

func testItem(data map[string]any) Item {
	return NewItem(data)
}

func baseItem() map[string]any {
	return map[string]any{
		"id":             123,
		"name":           "Test Item",
		"level":          float64(100),
		"is_stackable":   false,
		"is_equippable":  true,
		"item_class":     map[string]any{"name": "Armor"},
		"item_subclass":  map[string]any{"name": "Plate"},
		"inventory_type": "HEAD",
		"preview_item": map[string]any{
			"binding":      map[string]any{"type": "Binds on Equip"},
			"quality":      map[string]any{"name": "Epic"},
			"sell_price":   map[string]any{"value": float64(123456)},
			"requirements": map[string]any{"skill": map[string]any{"display_string": "Level 80"}},
		},
	}
}

func TestNewItem(t *testing.T) {
	before := time.Now()
	i := testItem(baseItem())
	after := time.Now()
	if i.ID() != 123 || i.Name() != "Test Item" {
		t.Fatalf("unexpected item: %#v", i)
	}
	if i.Updated().Before(before) || i.Updated().After(after) {
		t.Fatalf("unexpected update time: %v", i.Updated())
	}
}

func TestItemAccessors(t *testing.T) {
	i := testItem(baseItem())
	checks := map[string]string{
		"Binding": i.Binding(), "InventoryType": i.InventoryType(), "Subclass": i.ItemSubclassName(),
		"Class": i.ItemClassName(), "Quality": i.Quality(), "Requirements": i.Requirements(),
	}
	want := map[string]string{"Binding": "Binds on Equip", "InventoryType": "HEAD", "Subclass": "Plate", "Class": "Armor", "Quality": "Epic", "Requirements": "Level 80"}
	for k, got := range checks {
		if got != want[k] {
			t.Errorf("%s=%q want %q", k, got, want[k])
		}
	}
	if i.ItemLevel() != 100 {
		t.Errorf("Item level: want %d, got %d", 100, i.ItemLevel())
	}
	if i.SellPriceAdvertised() != 123456 {
		t.Errorf("Sell price advertised: want %d, got %d", 123456, i.SellPriceAdvertised())
	}
	if i.SellPriceRealizable() != 0 {
		// This item is ARMOR. We don't know how to price that, so we assign zero.
		t.Errorf("Sell price realizable: want %d, got %d", 0, i.SellPriceRealizable())
	}

	if !i.Equippable() {
		t.Errorf("Item is not equippable, expected it to be equippable")
	}
	if i.Stackable() {
		t.Errorf("Item is stackable, expected it to not be stackable")
	}
	if !i.VariableItemLevel() {
		t.Errorf("Item has does not have variable item level, expected it to have variable item level")
	}
	if i.Toy() {
		t.Errorf("Item is toy, expected it to not be toy")
	}
}

func TestEquippableFallback(t *testing.T) {
	for _, typ := range []string{"HEAD", "ONE_HANDED", "SHIELD"} {
		data := baseItem()
		delete(data, "is_equippable")
		data["inventory_type"] = typ
		if !testItem(data).Equippable() {
			t.Errorf("%s should be equippable", typ)
		}
	}
	data := baseItem()
	delete(data, "is_equippable")
	data["inventory_type"] = "BAG"
	if testItem(data).Equippable() {
		t.Error("BAG should not be equippable")
	}
}

func TestEquippableAuthoritativeField(t *testing.T) {
	data := baseItem()
	data["is_equippable"] = false
	data["inventory_type"] = "HEAD"
	if testItem(data).Equippable() {
		t.Error("explicit false should win")
	}
}

func TestMissingOptionalFields(t *testing.T) {
	data := baseItem()
	delete(data, "preview_item")
	delete(data, "inventory_type")
	i := testItem(data)
	if i.Binding() != "" || i.InventoryType() != "UNKNOWN" || i.RelicType() != "" || i.Quality() != "" || i.Requirements() != "" || i.SellPriceAdvertised() != 0 || i.Toy() {
		t.Error("missing optional fields not handled")
	}
}

func TestCosmetic(t *testing.T) {
	cases := []struct {
		name    string
		class   string
		level   float64
		quality string
		want    bool
	}{
		{"explicit", "Armor", 100, "Common", true},
		{"rare level one", "Armor", 1, "Rare", true},
		{"epic level one", "Weapon", 1, "Epic", true},
		{"rare higher", "Armor", 2, "Rare", false},
		{"other class", "Consumable", 1, "Rare", false},
	}
	for _, tc := range cases {
		data := baseItem()
		data["item_class"] = map[string]any{"name": tc.class}
		data["level"] = tc.level
		data["preview_item"].(map[string]any)["quality"] = map[string]any{"name": tc.quality}
		if tc.name == "explicit" {
			data["item_subclass"] = map[string]any{"name": "Cosmetic"}
		}
		if got := testItem(data).Cosmetic(); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestAppearances(t *testing.T) {
	data := baseItem()
	data["appearances"] = []any{map[string]any{"id": float64(10)}, map[string]any{"id": float64(20)}}
	got := testItem(data).Appearances()
	if len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Fatalf("appearances=%v", got)
	}
	delete(data, "appearances")
	if testItem(data).Appearances() != nil {
		t.Error("missing appearances should be nil")
	}
}

func TestSellPriceRealizableVariableLevel(t *testing.T) {
	data := baseItem()
	data["level"] = float64(1)
	if testItem(data).SellPriceRealizable() != 0 {
		t.Error("variable-level armor should have no realizable price")
	}
}

func TestStale(t *testing.T) {
	i := testItem(baseItem())
	i.XUpdated = time.Now().Add(-48 * time.Hour)
	if !i.Stale(24 * time.Hour) {
		t.Error("old item should be stale")
	}
	if i.Stale(72 * time.Hour) {
		t.Error("recent-enough item should not be stale")
	}
}
