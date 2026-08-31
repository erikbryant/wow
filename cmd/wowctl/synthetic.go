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

func consumable(item *syntheticitem.Item) {
	item.SetItemClassName("Consumable")
}

func l20(item *syntheticitem.Item) {
	item.SetItemLevel(20)
}

func l21(item *syntheticitem.Item) {
	item.SetItemLevel(21)
}

func l23(item *syntheticitem.Item) {
	item.SetItemLevel(23)
}

func price(g, s, c int64) syntheticOption {
	return func(item *syntheticitem.Item) {
		item.SetPreviewPrice(common.Coppers(g, s, c))
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

		newItem(217958, "Used Socks", price(0, 0, 1)),
		newItem(217962, "Dud Bomb", price(0, 0, 1)),
		newItem(226001, "Pure Gold Stein", price(200, 0, 0), l23),
		newItem(226002, "Expensive-Looking Find", price(200, 0, 0), l23),
		newItem(226004, "Olden Text", price(200, 0, 0), l23),
		newItem(226005, "Ancient Tool", price(200, 0, 0), l23),

		// Items WITHOUT vendor prices and are Commodities.

		newItem(178149, "Centurion Anima Core", commodity),
		newItem(225784, "Potion of Polymorphic Translation: Nerubian", commodity),

		// Items WITH vendor prices and are Commodities.

		newItem(23704, "Eversong Port", price(0, 0, 75), commodity, consumable),
		newItem(43557, "Poisonous Ivy Berries", price(0, 0, 25), commodity, l20),
		newItem(54629, "Prickly Thorn", price(0, 0, 43), commodity),
		newItem(60390, "Reticulated Tissue", price(0, 19, 73), commodity),
		newItem(60405, "Stubby Bear Tail", price(0, 22, 22), commodity),
		newItem(60406, "Blood-Caked Incisors", price(0, 37, 27), commodity),
		newItem(62370, "Bear Whisker", price(0, 0, 1), commodity),
		newItem(62770, "Infested Feather", price(0, 0, 3), commodity),
		newItem(201420, "Gnolan's House Special", price(0, 18, 75), commodity, l21, consumable),
		newItem(201421, "Tuskarr Jerky", price(0, 12, 50), commodity, consumable),
		newItem(204836, "Insect Treasure", price(0, 0, 50), commodity, l21),
		newItem(204837, "Rotting Fruit", price(0, 0, 50), commodity, l21),
		newItem(204838, "Discarded Toy", price(0, 0, 50), commodity, l21),
		newItem(204840, "Bottled Pheromones", price(0, 0, 50), commodity, l21),
		newItem(204842, "Red Sparklepretty", price(0, 0, 50), commodity, l21),
		newItem(212531, "Ruined Candle", price(50, 0, 0), commodity),
		newItem(212533, "Ear Worm", price(50, 0, 0), commodity),
		newItem(212534, "Wax Carving of a Candle", price(50, 0, 0), commodity),
		newItem(213234, "Rusty Ritual Knife", price(20, 0, 0), commodity, l23),
		newItem(213235, "Summoning Circle Chalk", price(10, 0, 0), commodity, l23),
		newItem(213237, "Harbinger Idol", price(20, 0, 0), commodity, l23),
		newItem(213238, "Broken Shadow Beast Binding", price(10, 0, 0), commodity, l23),
		newItem(213240, "Decorated Truffle", price(30, 0, 0), commodity, l23),
		newItem(213241, "Gibbering Glowcap", price(10, 0, 0), commodity, l23),
		newItem(213242, "Adventures of Libarbie and Lichen", price(30, 0, 0), commodity, l23),
		newItem(213245, "Gnawed Binding", price(30, 0, 0), commodity, l23),
		newItem(213246, "Tiny Glowing Rock", price(30, 0, 0), commodity, l23),
		newItem(213247, "Razor-Sharp Bones", price(10, 0, 0), commodity, l23),
		newItem(213248, "Shredded Fish", price(10, 0, 0), commodity, l23),
		newItem(213249, "Empty Battered Shell", price(20, 0, 0), commodity, l23),
		newItem(213250, "Cracked Gem", price(10, 0, 0), commodity, l23),
		newItem(213251, "Cinderbee Wax Jar", price(20, 0, 0), commodity, l23),
		newItem(213252, "Stolen Earthen Contraption", price(10, 0, 0), commodity, l23),
		newItem(213253, "Gilded Candle", price(20, 0, 0), commodity, l23),
		newItem(213254, "Big Gold Nugget", price(10, 0, 0), commodity, l23),
		newItem(213255, "Wax Canary", price(20, 0, 0), commodity, l23),
		newItem(213256, "Wax Spoon", price(30, 0, 0), commodity, l23),
		newItem(213257, "Wax Shovel", price(30, 0, 0), commodity, l23),
		newItem(213258, "Odorant Oddity", price(10, 0, 0), commodity, l23),
		newItem(213259, "Silk Doll", price(20, 0, 0), commodity, l23),
		newItem(213260, "Ripped Shawl", price(20, 0, 0), commodity, l23),
		newItem(213261, "Niffen Smell Pouch", price(30, 0, 0), commodity, l23),
		newItem(213262, "Stained Glass Fragment", price(10, 0, 0), commodity, l23),
		newItem(213263, "Poison Needle", price(30, 0, 0), commodity, l23),
		newItem(213265, "Empty Antidote Vial", price(20, 0, 0), commodity, l23),
		newItem(213266, "Twitching Snack", price(10, 0, 0), commodity, l23),
		newItem(213267, "Idol of Ansurek", price(30, 0, 0), commodity, l23),
		newItem(213268, "Roughly Acquired Treat", price(20, 0, 0), commodity, l23),
		newItem(222906, "Plump Snapcrab", price(0, 0, 1), commodity),
		newItem(224153, "Nibbled Shroomcap", price(10, 0, 0), commodity, l23),
		newItem(224154, "Mushrock", price(20, 0, 0), commodity, l23),
		newItem(224155, "Peeled Fungal Scale", price(20, 0, 0), commodity, l23),
		newItem(225787, "Pheromone-Covered Missive", price(1, 87, 69), commodity, l23),
		newItem(225839, "Agitated Water", price(1, 98, 3), commodity, l23),
		newItem(228386, "Decaying Rope", price(0, 0, 1), commodity),
	}
}

func syntheticValidate(s []wowitem.Item, paths *path.Paths) error {
	err := wowapi.Init(paths.Secret)
	if err != nil {
		return err
	}

	client, err := wowapi.NewClient()
	if err != nil {
		return err
	}

	for _, item := range s {
		_, err := client.Item(strconv.FormatInt(item.ID(), 10))
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
