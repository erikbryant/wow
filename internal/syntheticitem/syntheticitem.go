package syntheticitem

import (
	"encoding/json"
	"strconv"
)

type Item struct {
	data map[string]any
}

func New(id int64) *Item {
	return &Item{
		data: map[string]any{
			"id": json.Number(strconv.FormatInt(id, 10)),
		},
	}
}

func (i *Item) Map() map[string]any {
	return i.data
}

func (i *Item) JSON() ([]byte, error) {
	return json.MarshalIndent(i.data, "", "  ")
}

func (i *Item) SetBinding(binding string) {
	i.set([]string{"preview_item", "binding", "type"}, binding)
}

func (i *Item) SetInventoryType(inventoryType string) {
	i.set([]string{"inventory_type", "type"}, inventoryType)
}

func (i *Item) SetLevel(level int64) {
	i.set([]string{"level"}, json.Number(strconv.FormatInt(level, 10)))
}

func (i *Item) SetItemSubclass(name string) {
	i.set([]string{"item_subclass", "name"}, name)
}

func (i *Item) SetItemClass(name string) {
	i.set([]string{"item_class", "name"}, name)
}

func (i *Item) SetStackable(stackable bool) {
	i.set([]string{"is_stackable"}, stackable)
}

func (i *Item) SetRelicType(relicType string) {
	i.set([]string{"preview_item", "gem_properties", "relic_type"}, relicType)
}

func (i *Item) SetSellPrice(value int64) {
	i.set([]string{"preview_item", "sell_price", "value"}, json.Number(strconv.FormatInt(value, 10)))
}

func (i *Item) SetRequiredSkill(displayString string) {
	i.set([]string{
		"preview_item",
		"requirements",
		"skill",
		"display_string",
	}, displayString)
}

func (i *Item) SetQuality(name string) {
	i.set([]string{"preview_item", "quality", "name"}, name)
}

func (i *Item) SetName(name string) {
	i.set([]string{"name"}, name)
}

func (i *Item) SetToy(toy bool) {
	i.set([]string{"preview_item", "toy"}, toy)
}

func (i *Item) SetAppearances(appearances []any) {
	i.set([]string{"appearances"}, appearances)
}

func (i *Item) set(keys []string, value any) {
	object := i.data

	for _, key := range keys[:len(keys)-1] {
		next, ok := object[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			object[key] = next
		}

		object = next
	}

	object[keys[len(keys)-1]] = value
}
