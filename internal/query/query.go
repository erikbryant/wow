package query

import "github.com/erikbryant/wow/internal/wowitem"

// Predicate is a function that determines whether an item matches a query.
type Predicate func(wowitem.Item) bool

func finder(items []wowitem.Item, predicates ...Predicate) []wowitem.Item {
	results := make([]wowitem.Item, 0)

	for _, item := range items {
		match := true

		for _, predicate := range predicates {
			if !predicate(item) {
				match = false
				break
			}
		}

		if match {
			results = append(results, item)
		}
	}

	return results
}

// Find returns all items matching every supplied predicate.
func Find(items []wowitem.Item, predicates ...Predicate) []wowitem.Item {
	return finder(items, predicates...)
}

// Count returns the number of items matching all supplied predicates.
func Count(items []wowitem.Item, predicates ...Predicate) int {
	// This could be made memory efficient by not allocating the slice.
	// But that leads to complexity. And this function isn't even used.
	return len(finder(items, predicates...))
}

// Any returns true if at least one item matches all supplied predicates.
func Any(items []wowitem.Item, predicates ...Predicate) bool {
	// This could be made faster by short-circuiting when the first is found.
	// But that leads to complexity. And this function isn't even used.
	return Count(items, predicates...) > 0
}
