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
