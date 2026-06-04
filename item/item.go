package item

import (
	"fmt"
	"log"
	"time"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/transmog"
)

// Item holds values about a WoW item
type Item struct {
	// WARNING: Changing this struct invalidates the cache.
	// The members have to be public to write to a gob file,
	// but only use the accessor functions!
	XId      int64
	XItem    map[string]interface{}
	XUpdated time.Time // Datetime when created or updated
}

// PetInfo contains the properties specific to a battle pet
type PetInfo struct {
	BreedId   int64
	Level     int64
	Name      string
	QualityId int64
	SpeciesId int64
}

// NewItem returns an Item populated with wowData
func NewItem(wowData map[string]interface{}) Item {
	return Item{
		XId:      web.ToInt64(wowData["id"]),
		XItem:    wowData,
		XUpdated: time.Now(),
	}
}

// Id returns the item ID
func (i Item) Id() int64 {
	return i.XId
}

// Binding returns whether and when the item binds
func (i Item) Binding() string {
	// The key is only sometimes there; do not error if it is missing
	value, _ := web.MsiValued(i.XItem, []string{"preview_item", "binding", "type"}, "")
	return value.(string)
}

// InventoryType returns the slot this item equips to, or UNKNOWN
func (i Item) InventoryType() string {
	// The key is only sometimes there; do not error if it is missing
	value, _ := web.MsiValued(i.XItem, []string{"inventory_type"}, "UNKNOWN")
	return value.(string)
}

// EquipSlotTypes is a lookup set for valid gear slots
var EquipSlotTypes = map[string]struct{}{
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

// Equippable returns true if the item is equippable
func (i Item) Equippable() bool {
	// Preferred authoritative field
	if v, ok := i.XItem["is_equippable"]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}

	// Fallback: inventory_type
	_, isEquipSlot := EquipSlotTypes[i.InventoryType()]
	return isEquipSlot
}

// ItemLevel returns the item level
func (i Item) ItemLevel() int64 {
	v, err := web.MsiValue(i.XItem, []string{"level"})
	if err != nil {
		log.Fatalf("Level: %s in %v", err, i.XItem)
	}
	return web.ToInt64(v)
}

// VariableItemLevel returns true if the item can be enhanced, changing its iLevel
func (i Item) VariableItemLevel() bool {
	if i.Stackable() {
		return false
	}
	cn := i.ItemClassName()
	return cn == "Armor" || cn == "Gem" || cn == "Weapon"
}

// ItemSubclassName returns the item subclass name
func (i Item) ItemSubclassName() string {
	v, _ := web.MsiValued(i.XItem, []string{"item_subclass", "name"}, "")
	return v.(string)
}

// Cosmetic returns true if this item is a cosmetic
func (i Item) Cosmetic() bool {
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
func (i Item) ItemClassName() string {
	v, err := web.MsiValue(i.XItem, []string{"item_class", "name"})
	if err != nil {
		log.Fatalf("ItemClassName: %s in %v", err, i.XItem)
	}
	return v.(string)
}

// Stackable returns true if the item can be stacked in the inventory
func (i Item) Stackable() bool {
	v, err := web.MsiValue(i.XItem, []string{"is_stackable"})
	if err != nil {
		log.Fatalf("Stackable: %s in %v", err, i.XItem)
	}
	return v.(bool)
}

// RelicType returns the relic type
func (i Item) RelicType() string {
	// The key is only sometimes there; do not error if it is missing
	v, _ := web.MsiValued(i.XItem, []string{"preview_item", "gem_properties", "relic_type"}, "")
	return v.(string)
}

// Name returns the item name
func (i Item) Name() string {
	return i.XItem["name"].(string)
}

func (i Item) previewPrice() (int64, error) {
	v, err := web.MsiValue(i.XItem, []string{"preview_item", "sell_price", "value"})
	if err != nil {
		return 0, err
	}

	return web.ToInt64(v), nil
}

// SellPriceAdvertised returns the vendor sell price listed in the JSON
func (i Item) SellPriceAdvertised() int64 {
	pp, err := i.previewPrice()
	if err != nil {
		// Items with no preview price don't sell
		return 0
	}

	return pp
}

// SellPriceRealizable returns the actual price the vendor will offer for this specific item
func (i Item) SellPriceRealizable() int64 {
	if i.VariableItemLevel() {
		// I don't know how to price these
		return 0
	}
	return i.SellPriceAdvertised()
}

// Updated returns the last time this item was updated in the cache
func (i Item) Updated() time.Time {
	return i.XUpdated
}

func (i Item) Requirements() string {
	v, _ := web.MsiValued(i.XItem, []string{"preview_item", "requirements", "skill", "display_string"}, "")
	return web.ToString(v)
}

func (i Item) Quality() string {
	v, _ := web.MsiValued(i.XItem, []string{"preview_item", "quality", "name"}, "")
	return web.ToString(v)
}

// Stale returns whether the item is older than a given number of days
func (i Item) Stale(age time.Duration) bool {
	if age == 0 {
		return false
	}
	return time.Now().Sub(i.Updated()) > age
}

func (i Item) Toy() bool {
	v, _ := web.MsiValued(i.XItem, []string{"preview_item", "toy"}, "")
	return web.ToString(v) == "Toy"
}

// Appearances returns the appearance IDs this item provides
func (i Item) Appearances() []int64 {
	appearanceIds := []int64{}

	v, _ := web.MsiValued(i.XItem, []string{"appearances"}, nil)
	if v == nil {
		// Most items do not have appearances
		return nil
	}

	appearances := v.([]interface{})
	for _, appearance := range appearances {
		id := appearance.(map[string]interface{})["id"]
		appearanceIds = append(appearanceIds, web.ToInt64(id))
	}

	return appearanceIds
}

func (i Item) AppearanceSet() bool {
	return transmog.InAppearanceSet(i.Appearances())
}

// Format returns a formatted string representing the item
func (i Item) Format() string {
	equippable := "!E"
	if i.Equippable() {
		equippable = " E"
	}
	stackable := "!S"
	if i.Stackable() {
		stackable = " S"
	}
	appearanceSet := "!AS"
	if i.AppearanceSet() {
		appearanceSet = " AS"
	}
	return fmt.Sprintf("%7d  %s %s %s %11s   %3d   %-18s   %-8s   %s   %s", i.Id(), equippable, stackable, appearanceSet, common.Gold(i.SellPriceAdvertised()), i.ItemLevel(), i.ItemClassName(), i.Quality(), i.Updated().Format("2006-01-02"), i.Name())
}
