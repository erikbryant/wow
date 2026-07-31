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

// ByID sorts items by numeric item ID.
func ByID(a, b wowitem.Item) bool {
	return a.ID() < b.ID()
}
