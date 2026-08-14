package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/query"
	"github.com/erikbryant/wow/internal/wowitem"
)

func runQuery(args []string) error {
	flags := flag.NewFlagSet("query", flag.ExitOnError)

	rare := flags.Bool(
		"rare",
		false,
		"only rare items",
	)

	epic := flags.Bool(
		"epic",
		false,
		"only epic items",
	)

	inAppearanceSet := flags.Bool(
		"in-appearance-set",
		false,
		"only items in appearance sets",
	)

	class := flags.String(
		"class",
		"",
		"only items of this class",
	)

	name := flags.String(
		"name",
		"",
		"items whose name contains this text",
	)

	ilevelMin := flags.Int64(
		"ilevel-min",
		0,
		"minimum item level",
	)

	ilevelMax := flags.Int64(
		"ilevel-max",
		0,
		"maximum item level",
	)

	itemID := flags.Int64(
		"id",
		-1,
		"item ID",
	)

	sortField := flags.String(
		"sort",
		"id",
		"sort items by {id, name}",
	)

	if err := flags.Parse(args); err != nil {
		return err
	}

	var predicates []query.Predicate

	if *rare {
		predicates = append(predicates, query.Rare())
	}

	if *epic {
		predicates = append(predicates, query.Epic())
	}

	if *inAppearanceSet {
		predicates = append(predicates, query.AppearanceSet())
	}

	if *class != "" {
		predicates = append(predicates, query.ItemClass(*class))
	}

	if *name != "" {
		predicates = append(predicates, query.NameContains(*name))
	}

	if *ilevelMin != 0 {
		predicates = append(predicates, query.ItemLevelAtLeast(*ilevelMin))
	}

	if *ilevelMax != 0 {
		predicates = append(predicates, query.ItemLevelAtMost(*ilevelMax))
	}

	if *itemID != -1 {
		predicates = append(predicates, query.ItemID(*itemID))
	}

	wowItems := wowitem.New()
	items := wowItems.Values()

	results := query.Find(items, predicates...)

	switch *sortField {
	case "id":
		query.Sort(results, query.ByID)
	case "name":
		query.Sort(results, query.ByName)
	default:
		return fmt.Errorf("sort order must be one of {id, name} got %s", *sortField)
	}

	output.Table(os.Stdout, results)

	return nil
}
