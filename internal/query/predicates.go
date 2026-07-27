package query

import (
	"strings"

	"github.com/erikbryant/wow/internal/wowitem"
)

// AppearanceSet returns true for items that are in an appearance set.
func AppearanceSet(item wowitem.Item) bool {
	return item.AppearanceSet()
}

// IsRare returns true for rare-quality items.
func IsRare(item wowitem.Item) bool {
	return item.Quality() == "Rare"
}

// IsEpic returns true for epic-quality items.
func IsEpic(item wowitem.Item) bool {
	return item.Quality() == "Epic"
}

// IsArmor returns true for armor items.
func IsArmor(item wowitem.Item) bool {
	return item.ItemClassName() == "Armor"
}

// IsWeapon returns true for weapon items.
func IsWeapon(item wowitem.Item) bool {
	return item.ItemClassName() == "Weapon"
}

// NameContains returns true if the item name contains text.
func NameContains(text string) Predicate {
	text = strings.ToLower(text)

	return func(item wowitem.Item) bool {
		return strings.Contains(
			strings.ToLower(item.Name()),
			text,
		)
	}
}

// ItemLevelAtLeast returns true for items at or above an item level.
func ItemLevelAtLeast(level int64) Predicate {
	return func(item wowitem.Item) bool {
		return item.ItemLevel() >= level
	}
}

// ItemLevelAtMost returns true for items below or equal to an item level.
func ItemLevelAtMost(level int64) Predicate {
	return func(item wowitem.Item) bool {
		return item.ItemLevel() <= level
	}
}

// ItemClass returns true for items of a specific class.
func ItemClass(class string) Predicate {
	return func(item wowitem.Item) bool {
		return item.ItemClassName() == class
	}
}
