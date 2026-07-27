package query

import "github.com/erikbryant/wow/internal/wowitem"

// Predicate is a function that determines whether an item matches a query.
type Predicate func(wowitem.Item) bool

// Find returns all items matching every supplied predicate.
func Find(items []wowitem.Item, predicates ...Predicate) []wowitem.Item {
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

// Any returns true if at least one item matches all supplied predicates.
func Any(items []wowitem.Item, predicates ...Predicate) bool {
	for _, item := range items {
		match := true

		for _, predicate := range predicates {
			if !predicate(item) {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

// Count returns the number of items matching all supplied predicates.
func Count(items []wowitem.Item, predicates ...Predicate) int {
	count := 0

	for _, item := range items {
		match := true

		for _, predicate := range predicates {
			if !predicate(item) {
				match = false
				break
			}
		}

		if match {
			count++
		}
	}

	return count
}
