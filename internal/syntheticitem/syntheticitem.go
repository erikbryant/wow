package syntheticitem

import (
	"encoding/json"
	"strconv"
)

type Item struct {
	data map[string]any
}

func New(id int64) *Item {
	i := Item{
		data: map[string]any{},
	}

	i.SetID(id)
	i.SetStackable(false)

	return &i
}

func (i *Item) Map() map[string]any {
	return i.data
}

func (i *Item) JSON() ([]byte, error) {
	return json.MarshalIndent(i.data, "", "  ")
}

func (i *Item) SetID(id int64) *Item {
	i.set([]string{"id"}, json.Number(strconv.FormatInt(id, 10)))
	return i
}

func (i *Item) SetBinding(binding string) *Item {
	i.set([]string{"preview_item", "binding", "type"}, binding)
	return i
}

func (i *Item) SetInventoryType(inventoryType string) *Item {
	i.set([]string{"inventory_type", "type"}, inventoryType)
	return i
}

func (i *Item) SetLevel(level int64) *Item {
	i.set([]string{"level"}, json.Number(strconv.FormatInt(level, 10)))
	return i
}

func (i *Item) SetItemSubclassName(name string) *Item {
	i.set([]string{"item_subclass", "name"}, name)
	return i
}

func (i *Item) SetItemClassName(name string) *Item {
	i.set([]string{"item_class", "name"}, name)
	return i
}

func (i *Item) SetStackable(stackable bool) *Item {
	i.set([]string{"is_stackable"}, stackable)
	return i
}

func (i *Item) SetRelicType(relicType string) *Item {
	i.set([]string{"preview_item", "gem_properties", "relic_type"}, relicType)
	return i
}

func (i *Item) SetPreviewPrice(value int64) *Item {
	i.set([]string{"preview_item", "sell_price", "value"}, json.Number(strconv.FormatInt(value, 10)))
	return i
}

func (i *Item) SetRequiredSkill(displayString string) *Item {
	i.set([]string{
		"preview_item",
		"requirements",
		"skill",
		"display_string",
	}, displayString)
	return i
}

func (i *Item) SetQuality(name string) *Item {
	i.set([]string{"preview_item", "quality", "name"}, name)
	return i
}

func (i *Item) SetName(name string) *Item {
	i.set([]string{"name"}, name)
	return i
}

func (i *Item) SetToy(toy bool) *Item {
	i.set([]string{"preview_item", "toy"}, toy)
	return i
}

func (i *Item) SetAppearances(appearances []any) *Item {
	i.set([]string{"appearances"}, appearances)
	return i
}

func (i *Item) set(keys []string, value any) *Item {
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

	return i
}
