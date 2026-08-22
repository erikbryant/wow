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

type syntheticOption func(*syntheticitem.Item)

func commodity(item *syntheticitem.Item) {
	item.SetStackable(true)
}

func itemClassName(name string) syntheticOption {
	return func(item *syntheticitem.Item) {
		item.SetItemClassName(name)
	}
}

func price(g, s, c int64) syntheticOption {
	return func(item *syntheticitem.Item) {
		item.SetPreviewPrice(common.Coppers(g, s, c))
	}
}

func level(l int64) syntheticOption {
	return func(item *syntheticitem.Item) {
		item.SetItemLevel(l)
	}
}

// newItem returns a populated non-commodity Item.
func newItem(id int64, name string, options ...syntheticOption) wowitem.Item {
	item := syntheticitem.New(id, name)

	for _, option := range options {
		option(item)
	}

	return *wowitem.NewItem(item.Map())
}

// synthetics returns the synthetic items we have created.
func synthetics() []wowitem.Item {
	return []wowitem.Item{

		// Items WITHOUT vendor prices.

		newItem(123865, "Relic of Ursol"),
		newItem(123868, "Relic of Shakama"),
		newItem(123869, "Relic of Elune"),
		newItem(147455, "Water Stone"),
		newItem(203932, "Sentient Book"),
		newItem(217959, "Incomplete Painting"),
		newItem(225218, "Echoing Fragment: Hallowfall"),
		newItem(225219, "Echoing Fragment: The Ringing Deeps"),
		newItem(225236, "Echoing Fragment: Isle of Dorn"),
		newItem(225237, "Echoing Fragment: Azj-Kahet"),
		newItem(268944, "Souvenir Halazzi Idol"),
		newItem(268945, "Souvenir Nalorakk Mask"),
		newItem(268946, "Souvenir Jan'alai Key Chain"),
		newItem(268947, "Souvenir Akil'zon Shine Paper Weight"),
		newItem(268948, "Fine Antique Silvermoon Drapes"),
		newItem(268949, "Single Earthen Salt Shaker"),
		newItem(275670, "Bill of Lading"),

		// Items WITH vendor prices.

		newItem(217958, "Used Socks", level(1), price(0, 0, 1)),
		newItem(226001, "Pure Gold Stein", level(23), price(200, 0, 0)),
		newItem(226002, "Expensive-Looking Find", level(23), price(200, 0, 0)),
		newItem(226004, "Olden Text", level(23), price(200, 0, 0)),
		newItem(226005, "Ancient Tool", level(23), price(200, 0, 0)),

		// Items WITHOUT vendor prices and are Commodities.

		newItem(178149, "Centurion Anima Core", commodity),
		newItem(225784, "Potion of Polymorphic Translation: Nerubian", commodity),

		// Items WITH vendor prices and are Commodities.

		newItem(23704, "Eversong Port", level(1), price(0, 0, 75), commodity, itemClassName("Consumable")),
		newItem(43557, "Poisonous Ivy Berries", level(20), price(0, 0, 25), commodity),
		newItem(54629, "Prickly Thorn", level(1), price(0, 0, 43), commodity),
		newItem(60390, "Reticulated Tissue", level(1), price(0, 19, 73), commodity),
		newItem(60405, "Stubby Bear Tail", level(1), price(0, 22, 22), commodity),
		newItem(60406, "Blood-Caked Incisors", level(1), price(0, 37, 27), commodity),
		newItem(62770, "Infested Feather", level(1), price(0, 0, 3), commodity),
		newItem(201420, "Gnolan's House Special", level(21), price(0, 18, 75), commodity, itemClassName("Consumable")),
		newItem(201421, "Tuskarr Jerky", level(1), price(0, 12, 50), commodity, itemClassName("Consumable")),
		newItem(204836, "Insect Treasure", level(21), price(0, 0, 50), commodity),
		newItem(204837, "Rotting Fruit", level(21), price(0, 0, 50), commodity),
		newItem(204838, "Discarded Toy", level(21), price(0, 0, 50), commodity),
		newItem(204840, "Bottled Pheromones", level(21), price(0, 0, 50), commodity),
		newItem(204842, "Red Sparklepretty", level(21), price(0, 0, 50), commodity),
		newItem(212531, "Ruined Candle", level(1), price(50, 0, 0), commodity),
		newItem(212533, "Ear Worm", level(1), price(50, 0, 0), commodity),
		newItem(212534, "Wax Carving of a Candle", level(1), price(50, 0, 0), commodity),
		newItem(213234, "Rusty Ritual Knife", level(23), price(20, 0, 0), commodity),
		newItem(213235, "Summoning Circle Chalk", level(23), price(10, 0, 0), commodity),
		newItem(213237, "Harbinger Idol", level(23), price(20, 0, 0), commodity),
		newItem(213238, "Broken Shadow Beast Binding", level(23), price(10, 0, 0), commodity),
		newItem(213240, "Decorated Truffle", level(23), price(30, 0, 0), commodity),
		newItem(213242, "Adventures of Libarbie and Lichen", level(23), price(30, 0, 0), commodity),
		newItem(213245, "Gnawed Binding", level(23), price(30, 0, 0), commodity),
		newItem(213247, "Razor-Sharp Bones", level(23), price(10, 0, 0), commodity),
		newItem(213250, "Cracked Gem", level(23), price(10, 0, 0), commodity),
		newItem(213251, "Cinderbee Wax Jar", level(23), price(20, 0, 0), commodity),
		newItem(213252, "Stolen Earthen Contraption", level(23), price(10, 0, 0), commodity),
		newItem(213253, "Gilded Candle", level(23), price(20, 0, 0), commodity),
		newItem(213254, "Big Gold Nugget", level(23), price(10, 0, 0), commodity),
		newItem(213255, "Wax Canary", level(23), price(20, 0, 0), commodity),
		newItem(213256, "Wax Spoon", level(23), price(30, 0, 0), commodity),
		newItem(213257, "Wax Shovel", level(23), price(30, 0, 0), commodity),
		newItem(213258, "Odorant Oddity", level(23), price(10, 0, 0), commodity),
		newItem(213259, "Silk Doll", level(23), price(20, 0, 0), commodity),
		newItem(213261, "Niffen Smell Pouch", level(23), price(30, 0, 0), commodity),
		newItem(213262, "Stained Glass Fragment", level(23), price(10, 0, 0), commodity),
		newItem(213263, "Poison Needle", level(23), price(30, 0, 0), commodity),
		newItem(213266, "Twitching Snack", level(23), price(10, 0, 0), commodity),
		newItem(213267, "Idol of Ansurek", level(23), price(30, 0, 0), commodity),
		newItem(222906, "Plump Snapcrab", level(1), price(0, 0, 1), commodity),
		newItem(224153, "Nibbled Shroomcap", level(23), price(10, 0, 0), commodity),
		newItem(224154, "Mushrock", level(23), price(20, 0, 0), commodity),
		newItem(224155, "Peeled Fungal Scale", level(23), price(20, 0, 0), commodity),
		newItem(226003, "Snake Oil", level(23), price(200, 0, 0), commodity),
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
