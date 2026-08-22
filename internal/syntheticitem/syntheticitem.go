package syntheticitem

import (
	"encoding/json"
	"strconv"
)

type Item struct {
	data map[string]any
}

func New(id int64, name string) *Item {
	i := Item{
		data: map[string]any{},
	}

	i.SetID(id)
	i.SetName(name)
	i.SetItemLevel(1)
	i.SetStackable(false)
	i.SetItemClassName("Miscellaneous")

	return &i
}

func (i *Item) Map() map[string]any {
	return i.data
}

func (i *Item) SetID(id int64) *Item {
	i.set([]string{"id"}, json.Number(strconv.FormatInt(id, 10)))
	return i
}

func (i *Item) SetItemLevel(level int64) *Item {
	i.set([]string{"level"}, json.Number(strconv.FormatInt(level, 10)))
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

func (i *Item) SetPreviewPrice(value int64) *Item {
	i.set([]string{"preview_item", "sell_price", "value"}, json.Number(strconv.FormatInt(value, 10)))
	return i
}

func (i *Item) SetName(name string) *Item {
	i.set([]string{"name"}, name)
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
