package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/query"
	"github.com/erikbryant/wow/internal/syntheticitem"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// newItem returns a populated non-commodity Item.
func newItem(id int64, name string, level int64, sellPrice int64, class string) wowitem.Item {
	item := syntheticitem.New(id).
		SetName(name).
		SetLevel(level).
		SetPreviewPrice(sellPrice).
		SetItemClassName(class)

	return wowitem.NewItem(item.Map())
}

// newItemNoPrice returns a populated non-commodity Item with no sell price.
func newItemNoPrice(id int64, name string) wowitem.Item {
	item := syntheticitem.New(id).
		SetName(name).
		SetLevel(1).
		SetItemClassName("Miscellaneous")

	return wowitem.NewItem(item.Map())
}

// newCommodity returns a populated commodity Item.
func newCommodity(id int64, name string, level int64, sellPrice int64, class string) wowitem.Item {
	item := syntheticitem.New(id).
		SetName(name).
		SetLevel(level).
		SetPreviewPrice(sellPrice).
		SetItemClassName(class).
		SetStackable(true)

	return wowitem.NewItem(item.Map())
}

// newCommodityNoPrice returns a populated commodity Item with no sell price.
func newCommodityNoPrice(id int64, name string) wowitem.Item {
	item := syntheticitem.New(id).
		SetName(name).
		SetLevel(1).
		SetItemClassName("Miscellaneous").
		SetStackable(true)

	return wowitem.NewItem(item.Map())
}

// synthetics returns the synthetic items we have created.
func synthetics() []wowitem.Item {
	return []wowitem.Item{

		// Items that DO NOT have vendor prices. Just create the most basic placeholder item.

		newItemNoPrice(123865, "Relic of Ursol"),
		newItemNoPrice(123868, "Relic of Shakama"),
		newItemNoPrice(123869, "Relic of Elune"),
		newItemNoPrice(147455, "Water Stone"),
		newItemNoPrice(203932, "Sentient Book"),
		newItemNoPrice(217959, "Incomplete Painting"),
		newItemNoPrice(225218, "Echoing Fragment: Hallowfall"),
		newItemNoPrice(225219, "Echoing Fragment: The Ringing Deeps"),
		newItemNoPrice(225236, "Echoing Fragment: Isle of Dorn"),
		newItemNoPrice(225237, "Echoing Fragment: Azj-Kahet"),
		newItemNoPrice(268944, "Souvenir Halazzi Idol"),
		newItemNoPrice(268945, "Souvenir Nalorakk Mask"),
		newItemNoPrice(268946, "Souvenir Jan'alai Key Chain"),
		newItemNoPrice(268947, "Souvenir Akil'zon Shine Paper Weight"),
		newItemNoPrice(268948, "Fine Antique Silvermoon Drapes"),
		newItemNoPrice(268949, "Single Earthen Salt Shaker"),
		newItemNoPrice(275670, "Bill of Lading"),

		// Items that DO have vendor prices. These are more interesting.
		// They might actually become arbitrage plays.

		newItem(217958, "Used Socks", 1, common.Coppers(0, 0, 1), "Miscellaneous"),
		newItem(226001, "Pure Gold Stein", 23, common.Coppers(200, 0, 0), "Miscellaneous"),
		newItem(226002, "Expensive-Looking Find", 23, common.Coppers(200, 0, 0), "Miscellaneous"),
		newItem(226004, "Olden Text", 23, common.Coppers(200, 0, 0), "Miscellaneous"),
		newItem(226005, "Ancient Tool", 23, common.Coppers(200, 0, 0), "Miscellaneous"),

		// Commodities that DO NOT have vendor prices.

		newCommodityNoPrice(178149, "Centurion Anima Core"),
		newCommodityNoPrice(225784, "Potion of Polymorphic Translation: Nerubian"),

		// Commodities that DO have vendor prices.

		newCommodity(23704, "Eversong Port", 1, common.Coppers(0, 0, 75), "Consumable"),
		newCommodity(43557, "Poisonous Ivy Berries", 20, common.Coppers(0, 20, 25), "Miscellaneous"),
		newCommodity(54629, "Prickly Thorn", 1, common.Coppers(0, 0, 43), "Miscellaneous"),
		newCommodity(60390, "Reticulated Tissue", 1, common.Coppers(0, 19, 73), "Miscellaneous"),
		newCommodity(60405, "Stubby Bear Tail", 1, common.Coppers(0, 22, 22), "Miscellaneous"),
		newCommodity(60406, "Blood-Caked Incisors", 1, common.Coppers(0, 37, 27), "Miscellaneous"),
		newCommodity(62770, "Infested Feather", 1, common.Coppers(0, 0, 3), "Miscellaneous"),
		newCommodity(201420, "Gnolan's House Special", 21, common.Coppers(0, 18, 75), "Consumable"),
		newCommodity(201421, "Tuskarr Jerky", 1, common.Coppers(0, 12, 50), "Consumable"),
		newCommodity(204836, "Insect Treasure", 21, common.Coppers(0, 0, 50), "Miscellaneous"),
		newCommodity(204837, "Rotting Fruit", 21, common.Coppers(0, 0, 50), "Miscellaneous"),
		newCommodity(204838, "Discarded Toy", 21, common.Coppers(0, 0, 50), "Miscellaneous"),
		newCommodity(204840, "Bottled Pheromones", 21, common.Coppers(0, 0, 50), "Miscellaneous"),
		newCommodity(204842, "Red Sparklepretty", 21, common.Coppers(0, 0, 50), "Miscellaneous"),
		newCommodity(212531, "Ruined Candle", 1, common.Coppers(50, 0, 0), "Miscellaneous"),
		newCommodity(212533, "Ear Worm", 1, common.Coppers(50, 0, 0), "Miscellaneous"),
		newCommodity(212534, "Wax Carving of a Candle", 1, common.Coppers(50, 0, 0), "Miscellaneous"),
		newCommodity(213235, "Summoning Circle Chalk", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213238, "Broken Shadow Beast Binding", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213240, "Decorated Truffle", 23, common.Coppers(30, 0, 0), "Miscellaneous"),
		newCommodity(213242, "Adventures of Libarbie and Lichen", 23, common.Coppers(30, 0, 0), "Miscellaneous"),
		newCommodity(213245, "Gnawed Binding", 23, common.Coppers(30, 0, 0), "Miscellaneous"),
		newCommodity(213247, "Razor-Sharp Bones", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213250, "Cracked Gem", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213251, "Cinderbee Wax Jar", 23, common.Coppers(20, 0, 0), "Miscellaneous"),
		newCommodity(213252, "Stolen Earthen Contraption", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213253, "Gilded Candle", 23, common.Coppers(20, 0, 0), "Miscellaneous"),
		newCommodity(213254, "Big Gold Nugget", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213255, "Wax Canary", 23, common.Coppers(20, 0, 0), "Miscellaneous"),
		newCommodity(213258, "Odorant Oddity", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213259, "Silk Doll", 23, common.Coppers(20, 0, 0), "Miscellaneous"),
		newCommodity(213261, "Niffen Smell Pouch", 23, common.Coppers(30, 0, 0), "Miscellaneous"),
		newCommodity(213262, "Stained Glass Fragment", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(213263, "Poison Needle", 23, common.Coppers(30, 0, 0), "Miscellaneous"),
		newCommodity(213266, "Twitching Snack", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(224153, "Nibbled Shroomcap", 23, common.Coppers(10, 0, 0), "Miscellaneous"),
		newCommodity(224154, "Mushrock", 23, common.Coppers(20, 0, 0), "Miscellaneous"),
		newCommodity(224155, "Peeled Fungal Scale", 23, common.Coppers(20, 0, 0), "Miscellaneous"),
	}
}

func syntheticValidate(s []wowitem.Item, paths *path.Paths) error {
	err := wowapi.Init(paths.Secret)
	if err != nil {
		return err
	}

	for _, item := range s {
		_, err := wowapi.Item(strconv.FormatInt(item.ID(), 10))
		if err == nil {
			return fmt.Errorf("synthetic item with id %d shadows web API", item.ID())
		}
	}

	return nil
}

// syntheticList displays the synthetic items in a table.
func syntheticList(paths *path.Paths) error {
	as, err := appearanceset.New(paths.Appearances)
	if err != nil {
		return err
	}

	s := synthetics()
	query.Sort(s, query.ByID)
	output.Table(os.Stdout, s, as)

	return nil
}

// syntheticPopulate adds each of the synthetic items to the items persist
func syntheticPopulate(paths *path.Paths) error {
	s := synthetics()

	err := syntheticValidate(s, paths)
	if err != nil {
		return err
	}

	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

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
