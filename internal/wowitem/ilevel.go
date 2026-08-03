package wowitem

import "fmt"

var itemLevels = map[int64][]int64{
	// iLvl 70
	223969: {}, // Secret Sauce

	// iLvl 80
	237946: {180, 186, 192, 199, 206}, // Thalassian Needle Set
	237947: {180, 186, 192, 199, 206}, // Thalassian Leatherworker's Toolset
	237948: {180, 186, 192, 199, 206}, // Thalassian Blacksmith's Toolbox
	238009: {180, 186, 192, 199, 206}, // Thalassian Sickle
	238010: {},                        // Thalassian Pickaxe
	238011: {180, 186, 192, 199, 206}, // Thalassian Skinning Knife
	238012: {180, 186, 192, 199, 206}, // Thalassian Leatherworker's Knife
	238013: {180, 186, 192, 199, 206}, // Thalassian Blacksmith's Hammer
	239641: {180, 186, 192, 199, 206}, // Bright Linen Alchemy Apron
	239642: {180, 186, 192, 199, 206}, // Chef's Bright Linen Cooking Chapeau
	239643: {180, 186, 192, 199, 206}, // Bright Linen Enchanting Hat
	239645: {180, 186, 192, 199, 206}, // Bright Linen Herbalism Hat
	239646: {180, 186, 192, 199, 206}, // Bright Linen Tailoring Robe
	240953: {180, 186, 192, 199, 206}, // Bold Biographer's Bifocals
	240954: {180, 186, 192, 199, 206}, // Fantastic Font Focuser
	244175: {180, 186, 192, 199, 206}, // Runed Refulgent Copper Rod
	244615: {},                        // Eversong Botanist's Satchel
	244617: {},                        // Skinner's Cap
	244618: {180, 186, 192, 199, 206}, // Tinker's Handguard
	244619: {180, 186, 192, 199, 206}, // Hideworker's Cover
	244627: {180, 186, 192, 199, 206}, // Apprentice Smith's Apron
	244629: {180, 186, 192, 199, 206}, // Apprentice Jeweler's Apron
	244713: {},                        // Farstrider Clampers
	244717: {180, 186, 192, 199, 206}, // Junker's Multitool
	245775: {180, 186, 192, 199, 206}, // Hobbyist Scribe's Quill
	245777: {180, 186, 192, 199, 206}, // Hobbyist Alchemist's Mixing Rod
	245779: {},                        // Hobbyist Rolling Pin

	// iLvl 80 oddities
	240955: {180, 183, 186, 189, 193}, // Silvermoon Loupes
	240956: {180, 183, 186, 189, 193}, // Silvermoon Focusing Shard

	// iLvl 106
	237950: {},                        // Sun-Blessed Needle Set
	237951: {},                        // Sun-Blessed Leatherworker's Toolset
	237952: {206, 212, 218, 225, 232}, // Sun-Blessed Blacksmith's Toolbox
	238014: {},                        // Sun-Blessed Sickle
	238015: {},                        // Sun-Blessed Pickaxe
	238016: {},                        // Sun-Blessed Skinning Knife
	238017: {},                        // Sun-Blessed Leatherworker's Knife
	238018: {206, 212, 218, 225, 232}, // Sun-Blessed Blacksmith's Hammer
	239635: {},                        // Elegant Artisan's Alchemy Coveralls
	239636: {},                        // Elegant Artisan's Cooking Hat
	239637: {206, 212, 218, 225, 232}, // Elegant Artisan's Enchanting Hat
	239639: {},                        // Elegant Artisan's Herbalism Hat
	239640: {},                        // Elegant Artisan's Tailoring Robe
	240957: {},                        // Sin'dorei Scribe's Spectacles
	240958: {},                        // Improved Right-Handed Magnifying Glass
	240959: {206, 212, 218, 225, 232}, // Sin'dorei Jeweler's Loupes
	240960: {206, 212, 218, 225, 232}, // Sin'dorei Enchanter's Crystal
	244176: {},                        // Runed Brilliant Silver Rod
	244621: {},                        // Sin'dorei Herbalist's Backpack
	244622: {},                        // Sin'dorei Hunter's Pack
	244623: {},                        // Eversong Hunter's Headcover
	244624: {},                        // Sin'dorei Engineer's Gloves
	244625: {},                        // Sin'dorei Leathershaper's Smock
	244628: {206, 212, 218, 225, 232}, // Sin'dorei Forgemaster's Cover
	244630: {},                        // Sin'dorei Jeweler's Cover
	244714: {},                        // Sin'dorei Clampers
	244718: {206, 212, 218, 225, 232}, // Turbo-Junker's Multitool v1
	245776: {206, 212, 218, 225, 232}, // Sin'dorei Quill
	245778: {},                        // Sin'dorei Alchemist's Mixing Rod
	245780: {212, 218, 225, 232},      // Sin'dorei Rolling Pin

	// iLvl 106 oddities
	244616: {180, 186, 192, 199, 206}, // Skinner's Backpack

	// iLvl 270
	201601: {57, 58, 59, 61}, // Runed Serevite Rod

	// iLvl 317
	191233: {},                   // Chef's Smooth Rolling Pin
	191234: {},                   // Alchemist's Sturdy Mixing Rod
	191235: {70, 71, 72, 73, 74}, // Draconium Blacksmith's Toolbox
	191236: {71, 72, 73, 74},     // Draconium Leatherworker's Toolset
	191237: {70, 71, 72, 73, 74}, // Draconium Blacksmith's Hammer
	191238: {71, 72, 73, 74},     // Draconium Leatherworker's Knife
	191239: {71, 72, 73, 74},     // Draconium Needle Set
	191240: {70, 71, 72, 73, 74}, // Draconium Skinning Knife
	191241: {70, 71, 72, 73, 74}, // Draconium Sickle
	191242: {70, 71, 72, 73, 74}, // Draconium Pickaxe
	193479: {},                   // Floral Basket
	193480: {71, 72, 73, 74},     // Durable Pack
	193482: {},                   // Skinner's Cap
	193485: {},                   // Protective Gloves
	193486: {70, 71, 72, 73, 74}, // Resilient Smock
	193487: {71, 72, 73, 74},     // Alchemist's Hat
	193528: {},                   // Wildercloth Alchemist's Robe
	193534: {},                   // Wildercloth Chef's Hat
	193538: {},                   // Wildercloth Gardening Hat
	193539: {},                   // Wildercloth Enchanter's Hat
	193541: {70, 71, 72, 73, 74}, // Wildercloth Tailor's Coat
	193612: {71, 72, 73, 74},     // Smithing Apron
	193615: {71, 72, 73, 74},     // Jeweler's Cover
	194125: {},                   // Spring-Loaded Draconium Fabric Cutters
	194874: {71, 72, 74},         // Scribe's Fastened Quill
	198204: {},                   // Draconium Brainwave Amplifier
	198225: {},                   // Draconium Fisherfriend
	198234: {72, 74},             // Lapidary's Draconium Clamps
	198243: {},                   // Draconium Delver's Helmet
	198245: {},                   // Draconium Encased Samophlange
	198262: {},                   // Bottomless Stonecrust Ore Satchel
	198715: {70, 71, 72, 73, 74}, // Runed Draconium Rod

	// iLvl 317 oddities
	224114: {79, 85, 91, 98, 105}, // Runed Bismuth Rod

	// iLvl 486
	215117: {},                    // Storyteller's Glasses
	215119: {79, 85, 91, 98, 105}, // Right-Handed Magnifying Glass
	215120: {79, 85, 91, 98, 105}, // Radiant Loupes
	215121: {},                    // Incanter's Shard
	219861: {},                    // Gardener's Basket
	219862: {},                    // Hideseeker's Pack
	219863: {},                    // Hideseeker's Hat
	219864: {},                    // Scrapsmith's Gloves
	219865: {},                    // Hideshaper's Cover
	219866: {},                    // Apothecary's Cap
	219873: {},                    // Steelsmith's Apron
	219875: {},                    // Gemcutter's Apron
	221786: {},                    // Spring-Loaded Bismuth Fabric Cutters
	221788: {},                    // Bismuth Brainwave Projector
	221790: {85, 91, 98, 105},     // Bismuth Fisherfriend
	221792: {},                    // Lapidary's Bismuth Clamps
	221795: {},                    // Bismuth Miner's Headgear
	221797: {79, 85, 91, 98, 105}, // Bismuth-Fueled Samophlange
	221799: {},                    // Miner's Bismuth Hoard
	222480: {79, 85, 91, 98, 105}, // Proficient Sickle
	222481: {},                    // Proficient Pickaxe
	222482: {},                    // Proficient Skinning Knife
	222483: {79, 85, 91, 98, 105}, // Proficient Needle Set
	222484: {},                    // Proficient Leatherworker's Knife
	222485: {},                    // Proficient Leatherworker's Toolset
	222486: {},                    // Proficient Blacksmith's Hammer
	222487: {},                    // Proficient Blacksmith's Toolbox
	222573: {79, 85, 91, 98, 105}, // Lightweight Scribe's Quill
	222575: {79, 85, 91, 98, 105}, // Hasty Alchemist's Mixing Rod
	222577: {79, 85, 91, 98, 105}, // Burnt Rolling Pin
	222841: {},                    // Weavercloth Gardening Hat
	222843: {},                    // Weavercloth Enchanter's Hat
	222844: {},                    // Weavercloth Tailor's Coat
	222845: {},                    // Weavercloth Alchemist's Robe
	244711: {},                    // Farstrider Hobbyist Rod
	244715: {},                    // Farstrider Hardhat
	244719: {},                    // Farstrider Rock Satchel

	// iLvl 486 oddities
	222846: {79, 80, 81, 82, 83},      // Weavercloth Chef's Hat
	244620: {180, 186, 192, 199, 206}, // Chemist's Cap
	244707: {180, 186, 192, 199, 206}, // Farstrider Fabric Cutters
	244709: {180, 186, 192, 199, 206}, // Junker's Junk Visor

	// iLvl 535
	244626: {206, 212, 218, 225, 232}, // Sin'dorei Alchemist's Hat
	244708: {206, 212, 218, 225, 232}, // Sin'dorei Snippers
	244710: {},                        // Sin'dorei Headlamp
	244712: {},                        // Sin'dorei Angler's Rod
	244716: {},                        // Sin'dorei Gilded Hardhat
	244720: {},                        // Junker's Big Ol' Bag
}

// Known returns true if we have item level data for this itemID
func Known(itemID int64) bool {
	_, ok := itemLevels[itemID]
	return ok
}

// ILevels returns the item levels for this itemID
func ILevels(itemID int64) []int64 {
	iLevels, ok := itemLevels[itemID]
	if !ok {
		return []int64{0}
	}
	if len(iLevels) == 0 {
		fmt.Println("Missing item levels for:", itemID)
	}
	return iLevels
}
