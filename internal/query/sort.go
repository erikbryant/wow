package query

import (
	"sort"

	"github.com/erikbryant/wow/internal/wowitem"
)

// Sort sorts items in place using the supplied comparison function.
func Sort(items []wowitem.Item, less func(a, b wowitem.Item) bool) {
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})
}

// ByName sorts items alphabetically by name.
func ByName(a, b wowitem.Item) bool {
	return a.Name() < b.Name()
}

// ByItemLevel sorts items by item level ascending.
func ByItemLevel(a, b wowitem.Item) bool {
	return a.ItemLevel() < b.ItemLevel()
}

// ByQuality sorts items by quality ascending.
func ByQuality(a, b wowitem.Item) bool {
	return qualityRank(a) < qualityRank(b)
}

// ByID sorts items by numeric item ID.
func ByID(a, b wowitem.Item) bool {
	return a.ID() < b.ID()
}

// Reverse reverses any sort order.
func Reverse(less func(a, b wowitem.Item) bool) func(a, b wowitem.Item) bool {
	return func(a, b wowitem.Item) bool {
		return less(b, a)
	}
}

func qualityRank(item wowitem.Item) int {
	switch item.Quality() {
	case "Poor":
		return 0
	case "Common":
		return 1
	case "Uncommon":
		return 2
	case "Rare":
		return 3
	case "Epic":
		return 4
	case "Legendary":
		return 5
	case "Artifact":
		return 6
	case "Heirloom":
		return 7
	default:
		return -1
	}
}
