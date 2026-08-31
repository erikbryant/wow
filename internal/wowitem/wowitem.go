package wowitem

import (
	"fmt"
	"time"

	"github.com/erikbryant/wow/internal/common"
)

// Item holds values about a WoW item
type Item struct {
	// WARNING: Changing this struct invalidates the persistence,
	// even changing the variable names.
	// These members have to be public to write to a gob file,
	// but only use the accessor functions!
	XID      int64
	XItem    map[string]any
	XUpdated time.Time // Datetime when created or updated
}

// equipSlotTypes is a lookup set for valid gear slots
var equipSlotTypes = map[string]struct{}{
	"HEAD":       {},
	"NECK":       {},
	"SHOULDER":   {},
	"CHEST":      {},
	"WAIST":      {},
	"LEGS":       {},
	"FEET":       {},
	"WRIST":      {},
	"HANDS":      {},
	"FINGER":     {},
	"TRINKET":    {},
	"CLOAK":      {},
	"ONE_HANDED": {},
	"TWO_HANDED": {},
	"MAIN_HAND":  {},
	"OFF_HAND":   {},
	"RANGED":     {},
	"SHIELD":     {},
}

// NewItem returns an Item populated with wowData
func NewItem(wowData map[string]any) *Item {
	return &Item{
		XID:      common.JSONInt64Panic(wowData["id"]),
		XItem:    wowData,
		XUpdated: time.Now(),
	}
}

// ID returns the item ID
func (i *Item) ID() int64 {
	return i.XID
}

// Equippable returns true if the item is equippable
func (i *Item) Equippable() bool {
	v := common.MsaValued(i.XItem, []string{"is_equippable"}, false)
	return v.(bool)
}

// ItemLevel returns the item level
func (i *Item) ItemLevel() int64 {
	v, err := common.MsaValue(i.XItem, []string{"level"})
	if err != nil {
		panic(fmt.Errorf("level missing from %v: %w", i.XItem, err))
	}
	return common.JSONInt64Panic(v)
}

// VariableItemLevel returns true if the item can be enhanced, changing its ilevel
func (i *Item) VariableItemLevel() bool {
	if i.Stackable() {
		return false
	}
	cn := i.ItemClassName()
	return cn == "Armor" || cn == "Gem" || cn == "Weapon"
}

// ItemSubclassName returns the item subclass name
func (i *Item) ItemSubclassName() string {
	v := common.MsaValued(i.XItem, []string{"item_subclass", "name"}, "")
	return v.(string)
}

// Cosmetic returns true if this item is a cosmetic
func (i *Item) Cosmetic() bool {
	// Definitely cosmetic
	if i.ItemSubclassName() == "Cosmetic" {
		return true
	}

	// Likely to be cosmetic
	if i.ItemClassName() != "Armor" && i.ItemClassName() != "Weapon" {
		return false
	}
	if i.ItemLevel() == 1 && (i.Quality() == "Rare" || i.Quality() == "Epic") {
		return true
	}

	return false
}

// ItemClassName returns the item class name
func (i *Item) ItemClassName() string {
	v, err := common.MsaValue(i.XItem, []string{"item_class", "name"})
	if err != nil {
		panic(fmt.Errorf("item_class missing from %v: %w", i.XItem, err))
	}
	return v.(string)
}

// Stackable returns true if the item can be stacked in the inventory
func (i *Item) Stackable() bool {
	v, err := common.MsaValue(i.XItem, []string{"is_stackable"})
	if err != nil {
		panic(fmt.Errorf("is_stackable missing from %v: %w", i.XItem, err))
	}
	return v.(bool)
}

// Name returns the item name
func (i *Item) Name() string {
	return i.XItem["name"].(string)
}

func (i *Item) previewPrice() (int64, error) {
	v, err := common.MsaValue(i.XItem, []string{"preview_item", "sell_price", "value"})
	if err != nil {
		return 0, err
	}
	return common.JSONInt64Panic(v), nil
}

// SellPriceAdvertised returns the vendor sell price listed in the JSON
func (i *Item) SellPriceAdvertised() int64 {
	pp, err := i.previewPrice()
	if err != nil {
		// Items with no preview price don't sell
		return 0
	}
	return pp
}

// SellPriceRealizable returns the actual price the vendor will offer for this specific item
func (i *Item) SellPriceRealizable() int64 {
	if i.VariableItemLevel() {
		// I don't know how to price these
		return 0
	}
	return i.SellPriceAdvertised()
}

// Updated returns the last time this item was updated in the persistence
func (i *Item) Updated() time.Time {
	return i.XUpdated
}

// Quality returns the quality of this item
func (i *Item) Quality() string {
	v := common.MsaValued(i.XItem, []string{"preview_item", "quality", "name"}, "")
	return common.JSONString(v)
}

// Stale returns whether the item is older than a given number of days
func (i *Item) Stale(age time.Duration) bool {
	return time.Since(i.Updated()) > age
}

// Toy returns true if this item is a toy
func (i *Item) Toy() bool {
	v := common.MsaValued(i.XItem, []string{"preview_item", "toy"}, "")
	return common.JSONString(v) == "Toy"
}

// Appearances returns the appearance IDs this item provides
func (i *Item) Appearances() []int64 {
	appearanceIDs := []int64{}

	v := common.MsaValued(i.XItem, []string{"appearances"}, nil)
	if v == nil {
		// Most items do not have appearances
		return nil
	}

	appearances := v.([]any)
	for _, appearance := range appearances {
		id := appearance.(map[string]any)["id"]
		appearanceIDs = append(appearanceIDs, common.JSONInt64Panic(id))
	}

	return appearanceIDs
}
