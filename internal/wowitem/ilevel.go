package wowitem

import (
	"fmt"
	"os"
	"slices"
)

type Levels struct {
	want []int64
	has  []int64
}

var (
	iLvl70      = []int64{0}
	iLvl80      = []int64{180, 186, 192, 199, 206}
	iLvl80odd   = []int64{180, 183, 186, 189, 193}
	iLvl106     = []int64{206, 212, 218, 225, 232}
	iLvl106odd  = []int64{180, 186, 192, 199, 206}
	iLvl270     = []int64{57, 58, 59, 60, 61}
	iLvl317     = []int64{70, 71, 72, 73, 74}
	iLvl317odd  = []int64{79, 85, 91, 98, 105}
	iLvl486     = []int64{79, 85, 91, 98, 105}
	iLvl486odd1 = []int64{79, 80, 81, 82, 83}
	iLvl486odd2 = []int64{180, 186, 192, 199, 206}
	iLvl535     = []int64{206, 212, 218, 225, 232}
)

var itemLevels = map[int64]Levels{
	// iLvl 70
	223969: {iLvl70, iLvl70}, // Secret Sauce -- This does not appear to have distinct item levels

	// iLvl 80
	237946: {iLvl80, iLvl80},                      // Thalassian Needle Set
	237947: {iLvl80, iLvl80},                      // Thalassian Leatherworker's Toolset
	237948: {iLvl80, iLvl80},                      // Thalassian Blacksmith's Toolbox
	238009: {iLvl80, iLvl80},                      // Thalassian Sickle
	238010: {iLvl80, []int64{180, 186, 199, 206}}, // Thalassian Pickaxe
	238011: {iLvl80, iLvl80},                      // Thalassian Skinning Knife
	238012: {iLvl80, iLvl80},                      // Thalassian Leatherworker's Knife
	238013: {iLvl80, iLvl80},                      // Thalassian Blacksmith's Hammer
	239641: {iLvl80, iLvl80},                      // Bright Linen Alchemy Apron
	239642: {iLvl80, iLvl80},                      // Chef's Bright Linen Cooking Chapeau
	239643: {iLvl80, iLvl80},                      // Bright Linen Enchanting Hat
	239645: {iLvl80, iLvl80},                      // Bright Linen Herbalism Hat
	239646: {iLvl80, iLvl80},                      // Bright Linen Tailoring Robe
	240953: {iLvl80, iLvl80},                      // Bold Biographer's Bifocals
	240954: {iLvl80, iLvl80},                      // Fantastic Font Focuser
	244175: {iLvl80, iLvl80},                      // Runed Refulgent Copper Rod
	244615: {iLvl80, []int64{180, 186, 192, 206}}, // Eversong Botanist's Satchel
	244617: {iLvl80, iLvl80},                      // Skinner's Cap (sell price 2.50.00)
	244618: {iLvl80, iLvl80},                      // Tinker's Handguard
	244619: {iLvl80, iLvl80},                      // Hideworker's Cover
	244627: {iLvl80, iLvl80},                      // Apprentice Smith's Apron
	244629: {iLvl80, iLvl80},                      // Apprentice Jeweler's Apron
	244713: {iLvl80, iLvl80},                      // Farstrider Clampers
	244717: {iLvl80, iLvl80},                      // Junker's Multitool
	245775: {iLvl80, iLvl80},                      // Hobbyist Scribe's Quill
	245777: {iLvl80, iLvl80},                      // Hobbyist Alchemist's Mixing Rod
	245779: {iLvl80, []int64{180, 186, 192, 206}}, // Hobbyist Rolling Pin

	// iLvl 80 oddities
	240955: {iLvl80odd, iLvl80odd}, // Silvermoon Loupes
	240956: {iLvl80odd, iLvl80odd}, // Silvermoon Focusing Shard

	// iLvl 106
	237950: {iLvl106, []int64{218, 225, 232}},      // Sun-Blessed Needle Set
	237951: {iLvl106, []int64{225, 232}},           // Sun-Blessed Leatherworker's Toolset
	237952: {iLvl106, iLvl106},                     // Sun-Blessed Blacksmith's Toolbox
	238014: {iLvl106, []int64{218, 225, 232}},      // Sun-Blessed Sickle
	238015: {iLvl106, []int64{218, 225, 232}},      // Sun-Blessed Pickaxe
	238016: {iLvl106, []int64{218, 225, 232}},      // Sun-Blessed Skinning Knife
	238017: {iLvl106, []int64{212, 218, 225, 232}}, // Sun-Blessed Leatherworker's Knife
	238018: {iLvl106, iLvl106},                     // Sun-Blessed Blacksmith's Hammer
	239635: {iLvl106, []int64{212, 218, 225, 232}}, // Elegant Artisan's Alchemy Coveralls
	239636: {iLvl106, []int64{212, 218, 225, 232}}, // Elegant Artisan's Cooking Hat
	239637: {iLvl106, iLvl106},                     // Elegant Artisan's Enchanting Hat
	239639: {iLvl106, []int64{212, 218, 225, 232}}, // Elegant Artisan's Herbalism Hat
	239640: {iLvl106, []int64{218, 225, 232}},      // Elegant Artisan's Tailoring Robe
	240957: {iLvl106, iLvl106},                     // Sin'dorei Scribe's Spectacles
	240958: {iLvl106, iLvl106},                     // Improved Right-Handed Magnifying Glass
	240959: {iLvl106, iLvl106},                     // Sin'dorei Jeweler's Loupes
	240960: {iLvl106, iLvl106},                     // Sin'dorei Enchanter's Crystal
	244176: {iLvl106, []int64{218, 225, 232}},      // Runed Brilliant Silver Rod
	244621: {iLvl106, []int64{212, 225, 232}},      // Sin'dorei Herbalist's Backpack
	244622: {iLvl106, []int64{212, 218, 225, 232}}, // Sin'dorei Hunter's Pack
	244623: {iLvl106, []int64{218, 225, 232}},      // Eversong Hunter's Headcover
	244624: {iLvl106, []int64{212, 218, 225, 232}}, // Sin'dorei Engineer's Gloves
	244625: {iLvl106, iLvl106},                     // Sin'dorei Leathershaper's Smock
	244628: {iLvl106, iLvl106},                     // Sin'dorei Forgemaster's Cover
	244630: {iLvl106, []int64{212, 218, 232}},      // Sin'dorei Jeweler's Cover
	244714: {iLvl106, []int64{212, 225, 232}},      // Sin'dorei Clampers
	244718: {iLvl106, iLvl106},                     // Turbo-Junker's Multitool
	245776: {iLvl106, iLvl106},                     // Sin'dorei Quill
	245778: {iLvl106, []int64{212, 218, 225, 232}}, // Sin'dorei Alchemist's Mixing Rod
	245780: {iLvl106, []int64{212, 218, 225, 232}}, // Sin'dorei Rolling Pin

	// iLvl 106 oddities
	244616: {iLvl106odd, iLvl106odd}, // Skinner's Backpack

	// iLvl 270
	201601: {iLvl270, iLvl270}, // Runed Serevite Rod

	// iLvl 317
	191233: {iLvl317, []int64{71, 72, 74}},     // Chef's Smooth Rolling Pin
	191234: {iLvl317, []int64{71, 72, 74}},     // Alchemist's Sturdy Mixing Rod
	191235: {iLvl317, iLvl317},                 // Draconium Blacksmith's Toolbox
	191236: {iLvl317, []int64{71, 72, 73, 74}}, // Draconium Leatherworker's Toolset
	191237: {iLvl317, iLvl317},                 // Draconium Blacksmith's Hammer
	191238: {iLvl317, []int64{71, 72, 73, 74}}, // Draconium Leatherworker's Knife
	191239: {iLvl317, []int64{71, 72, 73, 74}}, // Draconium Needle Set
	191240: {iLvl317, iLvl317},                 // Draconium Skinning Knife
	191241: {iLvl317, iLvl317},                 // Draconium Sickle
	191242: {iLvl317, iLvl317},                 // Draconium Pickaxe
	193479: {iLvl317, []int64{71, 72, 73, 74}}, // Floral Basket
	193480: {iLvl317, []int64{71, 72, 73, 74}}, // Durable Pack
	193482: {iLvl317, []int64{71, 74}},         // Skinner's Cap (sell price 2.00.00)
	193485: {iLvl317, []int64{72, 73, 74}},     // Protective Gloves
	193486: {iLvl317, iLvl317},                 // Resilient Smock
	193487: {iLvl317, []int64{71, 72, 73, 74}}, // Alchemist's Hat
	193528: {iLvl317, []int64{71, 72, 73, 74}}, // Wildercloth Alchemist's Robe
	193534: {iLvl317, []int64{71, 72, 74}},     // Wildercloth Chef's Hat
	193538: {iLvl317, []int64{71, 72, 74}},     // Wildercloth Gardening Hat
	193539: {iLvl317, []int64{73, 74}},         // Wildercloth Enchanter's Hat
	193541: {iLvl317, iLvl317},                 // Wildercloth Tailor's Coat
	193612: {iLvl317, []int64{71, 72, 73, 74}}, // Smithing Apron
	193615: {iLvl317, []int64{71, 72, 73, 74}}, // Jeweler's Cover
	194125: {iLvl317, []int64{72, 74}},         // Spring-Loaded Draconium Fabric Cutters
	194874: {iLvl317, []int64{71, 72, 74}},     // Scribe's Fastened Quill
	198204: {iLvl317, []int64{72, 73}},         // Draconium Brainwave Amplifier
	198225: {iLvl317, []int64{72, 74}},         // Draconium Fisherfriend
	198234: {iLvl317, []int64{71, 72, 74}},     // Lapidary's Draconium Clamps
	198243: {iLvl317, []int64{72, 74}},         // Draconium Delver's Helmet
	198245: {iLvl317, []int64{71, 72, 74}},     // Draconium Encased Samophlange
	198262: {iLvl317, []int64{71, 72, 73, 74}}, // Bottomless Stonecrust Ore Satchel
	198715: {iLvl317, iLvl317},                 // Runed Draconium Rod

	// iLvl 317 oddities
	224114: {iLvl317odd, iLvl317odd}, // Runed Bismuth Rod

	// iLvl 486
	215117: {iLvl486, []int64{85, 91, 98, 105}}, // Storyteller's Glasses
	215119: {iLvl486, iLvl486},                  // Right-Handed Magnifying Glass
	215120: {iLvl486, iLvl486},                  // Radiant Loupes
	215121: {iLvl486, []int64{79, 85, 105}},     // Incanter's Shard
	219861: {iLvl486, []int64{85, 91, 105}},     // Gardener's Basket
	219862: {iLvl486, []int64{79, 85, 91, 105}}, // Hideseeker's Pack
	219863: {iLvl486, []int64{85, 91, 98, 105}}, // Hideseeker's Hat
	219864: {iLvl486, []int64{91, 98, 105}},     // Scrapsmith's Gloves
	219865: {iLvl486, []int64{85, 91, 105}},     // Hideshaper's Cover
	219866: {iLvl486, []int64{85, 105}},         // Apothecary's Cap
	219873: {iLvl486, []int64{85, 91, 98, 105}}, // Steelsmith's Apron
	219875: {iLvl486, []int64{85, 91, 98, 105}}, // Gemcutter's Apron
	221786: {iLvl486, []int64{98, 105}},         // Spring-Loaded Bismuth Fabric Cutters
	221788: {iLvl486, []int64{85, 98, 105}},     // Bismuth Brainwave Projector
	221790: {iLvl486, []int64{85, 91, 98, 105}}, // Bismuth Fisherfriend
	221792: {iLvl486, []int64{98, 105}},         // Lapidary's Bismuth Clamps
	221795: {iLvl486, []int64{91, 98, 105}},     // Bismuth Miner's Headgear
	221797: {iLvl486, iLvl486},                  // Bismuth-Fueled Samophlange
	221799: {iLvl486, []int64{91, 105}},         // Miner's Bismuth Hoard
	222480: {iLvl486, iLvl486},                  // Proficient Sickle
	222481: {iLvl486, []int64{79, 85, 98, 105}}, // Proficient Pickaxe
	222482: {iLvl486, []int64{91, 98, 105}},     // Proficient Skinning Knife
	222483: {iLvl486, iLvl486},                  // Proficient Needle Set
	222484: {iLvl486, []int64{79, 91, 98, 105}}, // Proficient Leatherworker's Knife
	222485: {iLvl486, []int64{85, 91, 98, 105}}, // Proficient Leatherworker's Toolset
	222486: {iLvl486, iLvl486},                  // Proficient Blacksmith's Hammer
	222487: {iLvl486, []int64{79, 85, 105}},     // Proficient Blacksmith's Toolbox
	222573: {iLvl486, iLvl486},                  // Lightweight Scribe's Quill
	222575: {iLvl486, iLvl486},                  // Hasty Alchemist's Mixing Rod
	222577: {iLvl486, iLvl486},                  // Burnt Rolling Pin
	244711: {iLvl486, []int64{}},                // Farstrider Hobbyist Rod

	// iLvl 486 oddities #1
	222841: {iLvl486odd1, []int64{81, 82, 83}},     // Weavercloth Gardening Hat
	222843: {iLvl486odd1, []int64{81, 82, 83}},     // Weavercloth Enchanter's Hat
	222844: {iLvl486odd1, []int64{80, 81, 82, 83}}, // Weavercloth Tailor's Coat
	222845: {iLvl486odd1, []int64{80, 81, 82, 83}}, // Weavercloth Alchemist's Robe
	222846: {iLvl486odd1, iLvl486odd1},             // Weavercloth Chef's Hat

	// iLvl 486 oddities #2
	244620: {iLvl486odd2, iLvl486odd2},                 // Chemist's Cap
	244707: {iLvl486odd2, iLvl486odd2},                 // Farstrider Fabric Cutters
	244709: {iLvl486odd2, iLvl486odd2},                 // Junker's Junk Visor
	244715: {iLvl486odd2, []int64{186, 192, 199, 206}}, // Farstrider Hardhat
	244719: {iLvl486odd2, iLvl486odd2},                 // Farstrider Rock Satchel

	// iLvl 535
	244626: {iLvl535, iLvl535},                     // Sin'dorei Alchemist's Hat
	244708: {iLvl535, iLvl535},                     // Sin'dorei Snippers
	244710: {iLvl535, []int64{212, 218, 225, 232}}, // Sin'dorei Headlamp
	244712: {iLvl535, []int64{212, 225, 232}},      // Sin'dorei Angler's Rod
	244716: {iLvl535, []int64{212, 225, 232}},      // Sin'dorei Gilded Hardhat
	244720: {iLvl535, []int64{225, 232}},           // Junker's Big Ol' Bag
}

// Known returns true if we have item level data for this itemID
func Known(itemID int64) bool {
	_, ok := itemLevels[itemID]
	return ok
}

// ILevels returns the item levels for this itemID
func ILevels(itemID int64) []int64 {
	levels, ok := itemLevels[itemID]
	if !ok {
		return []int64{0}
	}
	if len(levels.has) == 0 {
		fmt.Fprintf(os.Stderr, "*** missing item levels for: %d\n", itemID)
	}
	return levels.has
}

func missing(levels Levels) []int64 {
	m := map[int64]bool{}

	// Identify what we want
	for _, l := range levels.want {
		m[l] = true
	}

	// Eliminate the ones we have
	for _, l := range levels.has {
		_, ok := m[l]
		if !ok {
			panic(fmt.Sprintf("missing level %d", l))
		}
		delete(m, l)
	}

	// Return whatever is still needed
	need := []int64{}
	for l := range m {
		need = append(need, l)
	}

	return need
}

func ILevelsNeeded() []string {
	out := []string{}

	for id, levels := range itemLevels {
		need := missing(levels)
		for _, l := range need {
			out = append(out, fmt.Sprintf("    {%d, %d},", id, l))
		}
	}

	slices.Sort(out)

	return out
}
