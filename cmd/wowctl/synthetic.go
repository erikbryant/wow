package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/syntheticitem"
	"github.com/erikbryant/wow/internal/wowitem"
)

func synthetics() []wowitem.Item {
	var item *syntheticitem.Item
	items := []wowitem.Item{}

	item = syntheticitem.New(226002)
	item.SetName("Expensive-Looking Find")
	item.SetSellPrice(common.Coppers(1000000, 0, 0))
	item.SetStackable(false)
	item.SetLevel(0)
	item.SetItemClass("foobar")
	items = append(items, wowitem.NewItem(item.Map()))

	return items
}

func syntheticList(paths *path.Paths) error {
	as, err := appearanceset.New(paths.Appearances)
	if err != nil {
		return err
	}

	output.Table(os.Stdout, synthetics(), as)

	return nil
}

func syntheticPopulate(paths *path.Paths) error {
	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

	s := synthetics()
	for _, item := range s {
		wowItems.Set(item.ID(), item)
	}

	err = wowItems.Save()
	if err != nil {
		return fmt.Errorf("failed to save item persistence: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Populated %d items:\n\n", len(s))
	err = syntheticList(paths)
	if err != nil {
		return err
	}

	return nil
}

func runSynthetic(args []string, paths *path.Paths) error {
	if len(args) != 1 {
		usage()
		return fmt.Errorf("synthetic requires one argument")
	}

	cmd := args[0]

	switch cmd {
	case "list":
		return syntheticList(paths)
	case "populate":
		return syntheticPopulate(paths)
	default:
		usage()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
