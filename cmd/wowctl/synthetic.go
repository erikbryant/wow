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

func newWidget(id int64, name string) wowitem.Item {
	item := syntheticitem.New(id).
		SetName(name).
		SetLevel(1).
		SetItemClassName("Miscellaneous")

	return wowitem.NewItem(item.Map())
}

// synthetics returns the synthetic items we have created.
func synthetics() []wowitem.Item {
	var item *syntheticitem.Item
	items := []wowitem.Item{}

	// Items that DO NOT have vendor prices. Just create a lightweight entry.

	items = append(items, newWidget(123865, "Relic of Ursol"))
	items = append(items, newWidget(123865, "Relic of Ursol"))
	items = append(items, newWidget(123868, "Relic of Shakama"))
	items = append(items, newWidget(123869, "Relic of Elune"))
	items = append(items, newWidget(147455, "Water Stone"))
	items = append(items, newWidget(203932, "Sentient Book"))
	items = append(items, newWidget(217959, "Incomplete Painting"))
	items = append(items, newWidget(225218, "Echoing Fragment: Hallowfall"))
	items = append(items, newWidget(225219, "Echoing Fragment: The Ringing Deeps"))
	items = append(items, newWidget(225236, "Echoing Fragment: Isle of Dorn"))
	items = append(items, newWidget(225237, "Echoing Fragment: Azj-Kahet"))
	items = append(items, newWidget(268944, "Souvenir Halazzi Idol"))
	items = append(items, newWidget(268945, "Souvenir Nalorakk Mask"))
	items = append(items, newWidget(268946, "Souvenir Jan'alai Key Chain"))
	items = append(items, newWidget(268947, "Souvenir Akil'zon Shine Paper Weight"))
	items = append(items, newWidget(268948, "Fine Antique Silvermoon Drapes"))
	items = append(items, newWidget(268949, "Single Earthen Salt Shaker"))
	items = append(items, newWidget(275670, "Bill of Lading"))

	// Items that DO have vendor prices. These are much more interesting.
	// They might actually become arbitrage plays.

	item = syntheticitem.New(217958).
		SetName("Used Socks").
		SetLevel(1).
		SetPreviewPrice(common.Coppers(0, 0, 1)).
		SetItemClassName("Miscellaneous")
	items = append(items, wowitem.NewItem(item.Map()))

	item = syntheticitem.New(226002).
		SetName("Expensive-Looking Find").
		SetLevel(23).
		SetPreviewPrice(common.Coppers(200, 0, 0)).
		SetItemClassName("Miscellaneous")
	items = append(items, wowitem.NewItem(item.Map()))

	item = syntheticitem.New(226004).
		SetName("Olden Text").
		SetLevel(23).
		SetPreviewPrice(common.Coppers(200, 0, 0)).
		SetItemClassName("Miscellaneous")
	items = append(items, wowitem.NewItem(item.Map()))

	return items
}

// syntheticList displays the synthetic items in a table.
func syntheticList(paths *path.Paths) error {
	as, err := appearanceset.New(paths.Appearances)
	if err != nil {
		return err
	}

	output.Table(os.Stdout, synthetics(), as)

	return nil
}

// syntheticPopulate adds each of the synthetic items to the items persist
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

	fmt.Fprintf(os.Stdout, "\nPopulated %d items:\n\n", len(s))
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
