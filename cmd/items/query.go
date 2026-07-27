package main

import (
	"flag"
	"os"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/query"
)

func runQuery(args []string) {
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

	jsonOutput := flags.Bool(
		"json",
		false,
		"output JSON",
	)

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	var predicates []query.Predicate

	if *rare {
		predicates = append(predicates, query.IsRare)
	}

	if *epic {
		predicates = append(predicates, query.IsEpic)
	}

	if *inAppearanceSet {
		predicates = append(predicates, query.AppearanceSet)
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

	items := itemcache.ItemsSlice()

	results := query.Find(items, predicates...)

	if *jsonOutput {
		output.JSON(os.Stdout, results)
		return
	}

	output.Table(os.Stdout, results)
}
