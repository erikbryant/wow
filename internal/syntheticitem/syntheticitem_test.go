package syntheticitem

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/erikbryant/wow/internal/wowitem"
)

func TestNew(t *testing.T) {
	item := New(123456, "test name")

	if item == nil {
		t.Fatal("New returned nil")
	}

	if item.Map() == nil {
		t.Fatal("New returned item with nil map")
	}

	if got := item.Map()["id"]; got != json.Number("123456") {
		t.Errorf("id = %#v, want json.Number(\"123456\")", got)
	}

	if got := item.Map()["is_stackable"]; got != false {
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
			item := New(want, "test name")

			if got := item.Map()["id"]; got != json.Number(strconv.FormatInt(want, 10)) {
				t.Errorf("id = %#v, want %q", got, strconv.FormatInt(want, 10))
			}
		})
	}
}

func TestSetID(t *testing.T) {
	item := New(1, "test name").SetID(987654321)

	if got := item.Map()["id"]; got != json.Number("987654321") {
		t.Errorf("id = %#v, want json.Number(\"987654321\")", got)
	}
}

func TestSetLevel(t *testing.T) {
	item := New(1, "test name").SetItemLevel(123)

	if got := item.Map()["level"]; got != json.Number("123") {
		t.Errorf("level = %#v, want json.Number(\"123\")", got)
	}
}

func TestSetItemClassName(t *testing.T) {
	item := New(1, "test name").SetItemClassName("Armor")

	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Armor",
	)
}

func TestSetStackable(t *testing.T) {
	item := New(1, "test name")

	if got := item.Map()["is_stackable"]; got != false {
		t.Errorf("initial is_stackable = %#v, want false", got)
	}

	item.SetStackable(true)

	if got := item.Map()["is_stackable"]; got != true {
		t.Errorf("is_stackable = %#v, want true", got)
	}

	item.SetStackable(false)

	if got := item.Map()["is_stackable"]; got != false {
		t.Errorf("is_stackable = %#v, want false", got)
	}
}

func TestSetPreviewPrice(t *testing.T) {
	item := New(1, "test name").SetPreviewPrice(123456)

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "sell_price", "value"},
		json.Number("123456"),
	)
}

func TestSetName(t *testing.T) {
	item := New(1, "test name").SetName("Test Item")

	if got := item.Map()["name"]; got != "Test Item" {
		t.Errorf("name = %#v, want %q", got, "Test Item")
	}
}

func TestMutatorsAreChainable(t *testing.T) {
	item := New(123, "test name").
		SetName("Test Item").
		SetItemLevel(80).
		SetItemClassName("Armor").
		SetStackable(false).
		SetPreviewPrice(123456)

	if item == nil {
		t.Fatal("chained mutators returned nil")
	}

	assertNestedValue(t, item.Map(),
		[]string{"id"},
		json.Number("123"),
	)
	assertNestedValue(t, item.Map(),
		[]string{"name"},
		"Test Item",
	)
	assertNestedValue(t, item.Map(),
		[]string{"level"},
		json.Number("80"),
	)
	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Armor",
	)
	assertNestedValue(t, item.Map(),
		[]string{"is_stackable"},
		false,
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "sell_price", "value"},
		json.Number("123456"),
	)
}

func TestMutatorOverwritesExistingValue(t *testing.T) {
	item := New(1, "First").
		SetName("Second").
		SetItemLevel(10).
		SetItemLevel(20).
		SetItemClassName("Armor")

	if got := item.Map()["name"]; got != "Second" {
		t.Errorf("name = %#v, want %q", got, "Second")
	}

	if got := item.Map()["level"]; got != json.Number("20") {
		t.Errorf("level = %#v, want json.Number(\"20\")", got)
	}

	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Armor",
	)
}

func TestMutatorOverwritesNestedValue(t *testing.T) {
	item := New(1, "test name").
		SetItemClassName("Armor").
		SetPreviewPrice(100)

	item.SetItemClassName("Profession")
	item.SetPreviewPrice(200)

	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Profession",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "sell_price", "value"},
		json.Number("200"),
	)
}

func TestSetReplacesNonMapIntermediateValue(t *testing.T) {
	item := New(1, "test name")

	// Deliberately corrupt the intermediate structure. The setter should
	// replace it with a map rather than panic.
	item.Map()["preview_item"] = "not a map"

	item.SetItemClassName("Profession")

	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Profession",
	)
}

func TestMapReturnsUnderlyingMap(t *testing.T) {
	item := New(1, "test name")

	data := item.Map()
	data["name"] = "Externally Added"

	if got := item.Map()["name"]; got != "Externally Added" {
		t.Errorf("Map did not return underlying data map")
	}
}

func TestSyntheticIsItem(t *testing.T) {
	item := New(16, "test name").
		SetItemLevel(10).
		SetItemClassName("Armor").
		SetStackable(true).
		SetPreviewPrice(123456).
		SetName("test")

	wi := wowitem.NewItem(item.Map())

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

func assertNestedValue(t *testing.T, data map[string]any, keys []string, want any) {
	t.Helper()

	var current any = data

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
