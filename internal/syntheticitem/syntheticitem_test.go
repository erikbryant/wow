package syntheticitem

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	item := New(123456)

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
			item := New(want)

			if got := item.Map()["id"]; got != json.Number(strconv.FormatInt(want, 10)) {
				t.Errorf("id = %#v, want %q", got, strconv.FormatInt(want, 10))
			}
		})
	}
}

func TestSetID(t *testing.T) {
	item := New(1).SetID(987654321)

	if got := item.Map()["id"]; got != json.Number("987654321") {
		t.Errorf("id = %#v, want json.Number(\"987654321\")", got)
	}
}

func TestSetBinding(t *testing.T) {
	item := New(1).SetBinding("Binds on Equip")

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "binding", "type"},
		"Binds on Equip",
	)
}

func TestSetInventoryType(t *testing.T) {
	item := New(1).SetInventoryType("HEAD")

	got, ok := item.Map()["inventory_type"].(string)
	if !ok {
		t.Fatalf("inventory_type has type %T, want string", item.Map()["inventory_type"])
	}

	if got != "HEAD" {
		t.Errorf("inventory_type = %q, want %q", got, "HEAD")
	}
}

func TestSetLevel(t *testing.T) {
	item := New(1).SetLevel(123)

	if got := item.Map()["level"]; got != json.Number("123") {
		t.Errorf("level = %#v, want json.Number(\"123\")", got)
	}
}

func TestSetItemSubclassName(t *testing.T) {
	item := New(1).SetItemSubclassName("Plate")

	assertNestedValue(t, item.Map(),
		[]string{"item_subclass", "name"},
		"Plate",
	)
}

func TestSetItemClassName(t *testing.T) {
	item := New(1).SetItemClassName("Armor")

	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Armor",
	)
}

func TestSetStackable(t *testing.T) {
	item := New(1)

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

func TestSetRelicType(t *testing.T) {
	item := New(1).SetRelicType("Fire")

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "gem_properties", "relic_type"},
		"Fire",
	)
}

func TestSetPreviewPrice(t *testing.T) {
	item := New(1).SetPreviewPrice(123456)

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "sell_price", "value"},
		json.Number("123456"),
	)
}

func TestSetRequiredSkill(t *testing.T) {
	item := New(1).SetRequiredSkill("Level 80")

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "requirements", "skill", "display_string"},
		"Level 80",
	)
}

func TestSetQuality(t *testing.T) {
	item := New(1).SetQuality("Epic")

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "quality", "name"},
		"Epic",
	)
}

func TestSetName(t *testing.T) {
	item := New(1).SetName("Test Item")

	if got := item.Map()["name"]; got != "Test Item" {
		t.Errorf("name = %#v, want %q", got, "Test Item")
	}
}

func TestSetToy(t *testing.T) {
	item := New(1).SetToy(true)

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "toy"},
		"Toy",
	)

	item.SetToy(false)

	// SetToy(false) should retain the key but represent a non-toy value.
	if got := item.Map()["preview_item"].(map[string]any)["toy"]; got != "" {
		t.Errorf("toy after SetToy(false) = %#v, want \"\"", got)
	}
}

func TestSetAppearances(t *testing.T) {
	appearances := []any{
		map[string]any{"id": json.Number("100")},
		map[string]any{"id": json.Number("200")},
	}

	item := New(1).SetAppearances(appearances)

	got, ok := item.Map()["appearances"].([]any)
	if !ok {
		t.Fatalf("appearances has type %T, want []any", item.Map()["appearances"])
	}

	if !reflect.DeepEqual(got, appearances) {
		t.Errorf("appearances = %#v, want %#v", got, appearances)
	}
}

func TestMutatorsAreChainable(t *testing.T) {
	item := New(123).
		SetName("Test Item").
		SetLevel(80).
		SetInventoryType("HEAD").
		SetItemClassName("Armor").
		SetItemSubclassName("Plate").
		SetStackable(false).
		SetBinding("Binds on Equip").
		SetPreviewPrice(123456).
		SetRequiredSkill("Level 80").
		SetQuality("Epic").
		SetRelicType("Fire").
		SetToy(true).
		SetAppearances([]any{
			map[string]any{"id": json.Number("100")},
		})

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
		[]string{"inventory_type"},
		"HEAD",
	)
	assertNestedValue(t, item.Map(),
		[]string{"item_class", "name"},
		"Armor",
	)
	assertNestedValue(t, item.Map(),
		[]string{"item_subclass", "name"},
		"Plate",
	)
	assertNestedValue(t, item.Map(),
		[]string{"is_stackable"},
		false,
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "binding", "type"},
		"Binds on Equip",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "sell_price", "value"},
		json.Number("123456"),
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "requirements", "skill", "display_string"},
		"Level 80",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "quality", "name"},
		"Epic",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "gem_properties", "relic_type"},
		"Fire",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "toy"},
		"Toy",
	)
}

func TestMutatorOverwritesExistingValue(t *testing.T) {
	item := New(1).
		SetName("First").
		SetName("Second").
		SetLevel(10).
		SetLevel(20).
		SetQuality("Common").
		SetQuality("Epic")

	if got := item.Map()["name"]; got != "Second" {
		t.Errorf("name = %#v, want %q", got, "Second")
	}

	if got := item.Map()["level"]; got != json.Number("20") {
		t.Errorf("level = %#v, want json.Number(\"20\")", got)
	}

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "quality", "name"},
		"Epic",
	)
}

func TestMutatorOverwritesNestedValue(t *testing.T) {
	item := New(1).
		SetBinding("Binds on Equip").
		SetQuality("Rare").
		SetPreviewPrice(100)

	item.SetBinding("Soulbound")
	item.SetQuality("Epic")
	item.SetPreviewPrice(200)

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "binding", "type"},
		"Soulbound",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "quality", "name"},
		"Epic",
	)
	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "sell_price", "value"},
		json.Number("200"),
	)
}

func TestSetReplacesNonMapIntermediateValue(t *testing.T) {
	item := New(1)

	// Deliberately corrupt the intermediate structure. The setter should
	// replace it with a map rather than panic.
	item.Map()["preview_item"] = "not a map"

	item.SetBinding("Binds on Equip")

	assertNestedValue(t, item.Map(),
		[]string{"preview_item", "binding", "type"},
		"Binds on Equip",
	)
}

func TestJSON(t *testing.T) {
	item := New(123).
		SetName("Test Item").
		SetLevel(80).
		SetInventoryType("HEAD").
		SetItemClassName("Armor").
		SetItemSubclassName("Plate").
		SetStackable(false).
		SetBinding("Binds on Equip").
		SetPreviewPrice(123456).
		SetRequiredSkill("Level 80").
		SetQuality("Epic").
		SetRelicType("Fire").
		SetToy(true).
		SetAppearances([]any{
			map[string]any{"id": json.Number("100")},
			map[string]any{"id": json.Number("200")},
		})

	data, err := item.JSON()
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON produced invalid JSON: %v", err)
	}

	if got := decoded["id"]; got != float64(123) {
		t.Errorf("decoded id = %#v, want 123", got)
	}

	if got := decoded["name"]; got != "Test Item" {
		t.Errorf("decoded name = %#v, want %q", got, "Test Item")
	}

	if got := decoded["inventory_type"]; got != "HEAD" {
		t.Errorf("decoded inventory_type = %#v, want %q", got, "HEAD")
	}

	preview, ok := decoded["preview_item"].(map[string]any)
	if !ok {
		t.Fatalf("preview_item has type %T, want map[string]any", decoded["preview_item"])
	}

	if got := preview["toy"]; got != "Toy" {
		t.Errorf("decoded toy = %#v, want %q", got, "Toy")
	}
}

func TestJSONPreservesLargeIntegers(t *testing.T) {
	const id int64 = 9223372036854775807

	item := New(id)

	data, err := item.JSON()
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var decoded map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("JSON produced invalid JSON: %v", err)
	}

	if got := decoded["id"]; got != json.Number("9223372036854775807") {
		t.Errorf("decoded id = %#v, want json.Number(\"9223372036854775807\")", got)
	}
}

func TestMapReturnsUnderlyingMap(t *testing.T) {
	item := New(1)

	data := item.Map()
	data["name"] = "Externally Added"

	if got := item.Map()["name"]; got != "Externally Added" {
		t.Errorf("Map did not return underlying data map")
	}
}

func TestEmptyJSON(t *testing.T) {
	item := &Item{data: map[string]any{}}

	data, err := item.JSON()
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON produced invalid JSON: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("decoded map has %d entries, want 0", len(decoded))
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

func mergeBase(got map[string]any, want map[string]any) map[string]any {
	// This helper exists only to make TestSetBinding focus on the structure
	// introduced by SetBinding without requiring the test to know every
	// default field created by New.
	result := make(map[string]any, len(got))
	for key, value := range got {
		result[key] = value
	}
	return result
}
