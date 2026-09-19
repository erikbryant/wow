package wowitem

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	item := NewSynthetic(123456, "test name")

	if item == nil {
		t.Fatal("New returned nil")
	}

	if item.XItem == nil {
		t.Fatal("New returned item with nil map")
	}

	if got := item.ID(); got != 123456 {
		t.Errorf("id = %#v, want json.Number(\"123456\")", got)
	}

	if got := item.Stackable(); got != false {
		t.Errorf("is_stackable = %#v, want false", got)
	}
}

func TestNewIDBoundaries(t *testing.T) {
	tests := []int64{
		0,
		1,
		-1,
		1<<31 - 1,
		1 << 31,
		1<<63 - 1,
		-1 << 63,
	}

	for _, want := range tests {
		t.Run(strconv.FormatInt(want, 10), func(t *testing.T) {
			item := NewSynthetic(want, "test name")

			if got := item.ID(); got != want {
				t.Errorf("id = %#v, want %q", got, strconv.FormatInt(want, 10))
			}
		})
	}
}

func TestSetID(t *testing.T) {
	item := NewSynthetic(1, "test name").SetID(987654321)

	if got := item.ID(); got != 987654321 {
		t.Errorf("id = %#v, want 987654321", got)
	}
}

func TestSetLevel(t *testing.T) {
	item := NewSynthetic(1, "test name").SetItemLevel(123)

	if got := item.ItemLevel(); got != 123 {
		t.Errorf("level = %#v, want json.Number(\"123\")", got)
	}
}

func TestSetItemClassName(t *testing.T) {
	item := NewSynthetic(1, "test name").SetItemClassName("Armor")

	assertNestedValue(t, item,
		[]string{"item_class", "name"},
		"Armor",
	)
}

func TestSetStackable(t *testing.T) {
	item := NewSynthetic(1, "test name")

	if got := item.Stackable(); got != false {
		t.Errorf("initial is_stackable = %#v, want false", got)
	}

	item.SetStackable(true)

	if got := item.Stackable(); got != true {
		t.Errorf("is_stackable = %#v, want true", got)
	}

	item.SetStackable(false)

	if got := item.Stackable(); got != false {
		t.Errorf("is_stackable = %#v, want false", got)
	}
}

func TestSetPreviewPrice(t *testing.T) {
	item := NewSynthetic(1, "test name").SetPreviewPrice(123456)

	assertNestedValue(t, item,
		[]string{"preview_item", "sell_price", "value"},
		json.Number("123456"),
	)
}

func TestSetName(t *testing.T) {
	item := NewSynthetic(1, "test name").SetName("Test Item")

	if got := item.Name(); got != "Test Item" {
		t.Errorf("name = %#v, want %q", got, "Test Item")
	}
}

func TestMutatorsAreChainable(t *testing.T) {
	item := NewSynthetic(123, "test name").
		SetName("Test Item").
		SetItemLevel(80).
		SetItemClassName("Armor").
		SetStackable(false).
		SetPreviewPrice(123456)

	if item == nil {
		t.Fatal("chained mutators returned nil")
	}

	assertNestedValue(t, item,
		[]string{"id"},
		json.Number("123"),
	)
	assertNestedValue(t, item,
		[]string{"name"},
		"Test Item",
	)
	assertNestedValue(t, item,
		[]string{"level"},
		json.Number("80"),
	)
	assertNestedValue(t, item,
		[]string{"item_class", "name"},
		"Armor",
	)
	assertNestedValue(t, item,
		[]string{"is_stackable"},
		false,
	)
	assertNestedValue(t, item,
		[]string{"preview_item", "sell_price", "value"},
		json.Number("123456"),
	)
}

func TestMutatorOverwritesExistingValue(t *testing.T) {
	item := NewSynthetic(1, "First").
		SetName("Second").
		SetItemLevel(10).
		SetItemLevel(20).
		SetItemClassName("Armor")

	if got := item.Name(); got != "Second" {
		t.Errorf("name = %#v, want %q", got, "Second")
	}

	if got := item.ItemLevel(); got != 20 {
		t.Errorf("level = %#v, want json.Number(\"20\")", got)
	}

	assertNestedValue(t, item,
		[]string{"item_class", "name"},
		"Armor",
	)
}

func TestMutatorOverwritesNestedValue(t *testing.T) {
	item := NewSynthetic(1, "test name").
		SetItemClassName("Armor").
		SetPreviewPrice(100)

	item.SetItemClassName("Profession")
	item.SetPreviewPrice(200)

	assertNestedValue(t, item,
		[]string{"item_class", "name"},
		"Profession",
	)
	assertNestedValue(t, item,
		[]string{"preview_item", "sell_price", "value"},
		json.Number("200"),
	)
}

func TestSetReplacesNonMapIntermediateValue(t *testing.T) {
	item := NewSynthetic(1, "test name")

	// Deliberately corrupt the intermediate structure. The setter should
	// replace it with a map rather than panic.
	item.XItem["preview_item"] = "not a map"

	item.SetItemClassName("Profession")

	assertNestedValue(t, item,
		[]string{"item_class", "name"},
		"Profession",
	)
}

func TestMapReturnsUnderlyingMap(t *testing.T) {
	item := NewSynthetic(1, "test name")

	data := item.XItem
	data["name"] = "Externally Added"

	if got := item.Name(); got != "Externally Added" {
		t.Errorf("Map did not return underlying data map")
	}
}

func TestSyntheticIsItem(t *testing.T) {
	item := NewSynthetic(16, "test name").
		SetItemLevel(10).
		SetItemClassName("Armor").
		SetStackable(true).
		SetPreviewPrice(123456).
		SetName("test")

	wi := NewItem(item.XItem)

	if got := wi.ID(); got != 16 {
		t.Errorf("level = %#v, want %d", 16, got)
	}

	if got := wi.ItemLevel(); got != 10 {
		t.Errorf("level = %#v, want %d", 10, got)
	}

	if got := wi.ItemClassName(); got != "Armor" {
		t.Errorf("level = %#v, want %s", "Armor", got)
	}

	if got := wi.Stackable(); got != true {
		t.Errorf("level = %#v, want %t", true, got)
	}

	if got := wi.SellPriceAdvertised(); got != 123456 {
		t.Errorf("level = %#v, want %d", 123456, got)
	}

	if got := wi.Name(); got != "test" {
		t.Errorf("level = %#v, want %s", "test", got)
	}

}

func assertNestedValue(t *testing.T, i *Item, keys []string, want any) {
	t.Helper()

	var current any = i.XItem

	for _, key := range keys {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("while looking for %q: current value has type %T, want map[string]any", key, current)
		}

		current, ok = object[key]
		if !ok {
			t.Fatalf("missing key %q in %#v", key, object)
		}
	}

	if !reflect.DeepEqual(current, want) {
		t.Errorf("%v = %#v, want %#v", keys, current, want)
	}
}
