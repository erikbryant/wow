package item

var itemLevels = map[int64][]int64{
	// iLvl 80
	237946: {180, 186, 192, 199, 206}, // Thalassian Needle Set
	237947: {180, 186, 192, 199, 206}, // Thalassian Leatherworker's Toolset
	237948: {180, 186, 192, 199, 206}, // Thalassian Blacksmith's Toolbox
	238011: {180, 186, 192, 199, 206}, // Thalassian Skinning Knife
	238012: {180, 186, 192, 199, 206}, // Thalassian Leatherworker's Knife
	238013: {180, 186, 192, 199, 206}, // Thalassian Blacksmith's Hammer
	239641: {180, 186, 192, 199, 206}, // Bright Linen Alchemy Apron
	239642: {180, 186, 192, 199, 206}, // Chef's Bright Linen Cooking Chapeau
	239643: {180, 186, 192, 199, 206}, // Bright Linen Enchanting Hat
	239646: {180, 186, 192, 199, 206}, // Bright Linen Tailoring Robe
	240953: {180, 186, 192, 199, 206}, // Bold Biographer's Bifocals
	240954: {180, 186, 192, 199, 206}, // Fantastic Font Focuser
	244175: {180, 186, 192, 199, 206}, // Runed Refulgent Copper Rod
	244618: {180, 186, 192, 199, 206}, // Tinker's Handguard
	244619: {180, 186, 192, 199, 206}, // Hideworker's Cover
	244627: {180, 186, 192, 199, 206}, // Apprentice Smith's Apron
	244629: {180, 186, 192, 199, 206}, // Apprentice Jeweler's Apron
	245775: {180, 186, 192, 199, 206}, // Hobbyist Scribe's Quill
	245777: {180, 186, 192, 199, 206}, // Hobbyist Alchemist's Mixing Rod
	238009: {180, 186, 192, 199, 206}, // Thalassian Sickle
	244717: {180, 186, 192, 199, 206}, // Junker's Multitool

	// iLvl 80 oddities
	240955: {180, 183, 186, 189, 193}, // Silvermoon Loupes
	240956: {180, 183, 186, 189, 193}, // Silvermoon Focusing Shard

	// iLvl 106
	237952: {206, 212, 218, 225, 232}, // Sun-Blessed Blacksmith's Toolbox
	238018: {206, 212, 218, 225, 232}, // Sun-Blessed Blacksmith's Hammer
	240959: {206, 212, 218, 225, 232}, // Sin'dorei Jeweler's Loupes
	240960: {206, 212, 218, 225, 232}, // Sin'dorei Enchanter's Crystal
	244628: {206, 212, 218, 225, 232}, // Sin'dorei Forgemaster's Cover
	244718: {206, 212, 218, 225, 232}, // Turbo-Junker's Multitool v1
	245776: {206, 212, 218, 225, 232}, // Sin'dorei Quill
	245780: {212, 218, 225, 232},      // Sin'dorei Rolling Pin

	// iLvl 106 oddities
	244616: {180, 186, 192, 199, 206}, // Skinner's Backpack

	// iLvl 317
	191235: {70, 71, 72, 73, 74}, // Draconium Blacksmith's Toolbox
	191236: {71, 72, 73, 74},     // Draconium Leatherworker's Toolset
	191237: {70, 71, 72, 73, 74}, // Draconium Blacksmith's Hammer
	191238: {71, 72, 73, 74},     // Draconium Leatherworker's Knife
	191239: {71, 72, 73, 74},     // Draconium Needle Set
	191240: {70, 71, 72, 73, 74}, // Draconium Skinning Knife
	191241: {70, 71, 72, 73, 74}, // Draconium Sickle
	191242: {70, 71, 72, 73, 74}, // Draconium Pickaxe
	193486: {70, 71, 72, 73, 74}, // Resilient Smock
	193487: {71, 72, 73, 74},     // Alchemist's Hat
	193612: {71, 72, 73, 74},     // Smithing Apron
	193541: {70, 71, 72, 73, 74}, // Wildercloth Tailor's Coat
	198715: {70, 71, 72, 73, 74}, // Runed Draconium Rod

	// iLvl 317 oddities
	224114: {79, 85, 91, 98, 105}, // Runed Bismuth Rod

	// iLvl 486
	215119: {79, 85, 91, 98, 105}, // Right-Handed Magnifying Glass
	215120: {79, 85, 91, 98, 105}, // Radiant Loupes
	221797: {79, 85, 91, 98, 105}, // Bismuth-Fueled Samophlange
	222575: {79, 85, 91, 98, 105}, // Hasty Alchemist's Mixing Rod
	222577: {79, 85, 91, 98, 105}, // Burnt Rolling Pin
	222573: {79, 85, 91, 98, 105}, // Lightweight Scribe's Quill
	222483: {79, 85, 91, 98, 105}, // Proficient Needle Set

	// iLvl 486 oddities
	244709: {180, 186, 192, 199, 206}, // Junker's Junk Visor

	// iLvl 535
	244626: {206, 212, 218, 225, 232}, // Sin'dorei Alchemist's Hat
}

// Known returns true if the item has an entry in itemLevels
func Known(itemId int64) bool {
	_, ok := itemLevels[itemId]
	return ok
}

func ILevels(itemId, iLevel int64) []int64 {
	if Known(itemId) {
		return itemLevels[itemId]
	}

	// If there is no *specific* iLvl then use 0 to tell WoW we don't care
	return []int64{0}
}
