package query

import (
	"strings"

	"github.com/erikbryant/wow/internal/wowitem"
)

// AppearanceSet returns true for items that are in an appearance set.
func AppearanceSet() Predicate {
	return func(item wowitem.Item) bool {
		return item.AppearanceSet()
	}
}

// Rare returns true for rare-quality items.
func Rare() Predicate {
	return func(item wowitem.Item) bool {
		return item.Quality() == "Rare"
	}
}

// Epic returns true for epic-quality items.
func Epic() Predicate {
	return func(item wowitem.Item) bool {
		return item.Quality() == "Epic"
	}
}

// Armor returns true for armor items.
func Armor() Predicate {
	return func(item wowitem.Item) bool {
		return item.ItemClassName() == "Armor"
	}
}

// Weapon returns true for weapon items.
func Weapon() Predicate {
	return func(item wowitem.Item) bool {
		return item.ItemClassName() == "Weapon"
	}
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

// ItemID returns true for items with the given ID.
func ItemID(itemID int64) Predicate {
	return func(item wowitem.Item) bool {
		return item.ID() == itemID
	}
}
