package item

var itemLevels = map[int64][]int64{
	// iLvl 80
	237948: {180, 186, 192, 199, 206}, // Thalassian Blacksmith's Toolbox (Profession)
	238012: {180, 186, 192, 199, 206}, // Thalassian Leatherworker's Knife (Profession)
	239641: {180, 186, 192, 199, 206}, // Bright Linen Alchemy Apron (Profession)
	239646: {180, 186, 192, 199, 206}, // Bright Linen Tailoring Robe (Profession)
	240953: {180, 186, 192, 199, 206}, // Bold Biographer's Bifocals (Profession)
	240954: {180, 186, 192, 199, 206}, // Fantastic Font Focuser (Profession)
	240955: {180, 183, 186, 189, 193}, // Silvermoon Loupes (Profession)
	244175: {180, 186, 192, 199, 206}, // Runed Refulgent Copper Rod (Profession)
	244618: {180, 186, 192, 199, 206}, // Tinker's Handguard (Profession)
	244619: {180, 186, 192, 199, 206}, // Hideworker's Cover (Profession)
	244627: {180, 186, 192, 199, 206}, // Apprentice Smith's Apron (Profession)

	// iLvl 106
	237952: {206, 212, 218, 225, 232}, // Sun-Blessed Blacksmith's Toolbox (Profession)  iLvl: 106
	238018: {212, 218, 225, 232},      // Sun-Blessed Blacksmith's Hammer (Profession)
	240959: {206, 212, 218, 225, 232}, // Sin'dorei Jeweler's Loupes (Profession)
	244718: {206, 212, 218, 225, 232}, // Turbo-Junker's Multitool v1 (Profession)

	// iLvl 317
	191235: {71, 72, 73, 74},     // Draconium Blacksmith's Toolbox (Profession)
	191236: {71, 72, 73, 74},     // Draconium Leatherworker's Toolset (Profession)
	191237: {70, 71, 72, 73, 74}, // Draconium Blacksmith's Hammer (Profession)
	191238: {71, 72, 73, 74},     // Draconium Leatherworker's Knife (Profession)
	191239: {71, 72, 73, 74},     // Draconium Needle Set (Profession)
	191240: {71, 72, 73, 74},     // Draconium Skinning Knife (Profession)
	191241: {70, 71, 72, 74},     // Draconium Sickle (Profession)
	191242: {70, 71, 72, 74},     // Draconium Pickaxe (Profession)
	193486: {70, 71, 72, 74},     // Resilient Smock (Profession)
	224114: {79, 85, 91, 105},    // Runed Bismuth Rod (Profession)

	// iLvl 486
	215120: {79, 85, 91, 98, 105}, // Radiant Loupes (Profession)
	221797: {79, 85, 91, 98, 105}, // Bismuth-Fueled Samophlange (Profession)

	// iLvl 535
	244626: {206, 212, 218, 225, 232}, // Sin'dorei Alchemist's Hat (Profession)
}

// Known returns true if the item has an entry in itemLevels
func Known(itemId int64) bool {
	_, ok := itemLevels[itemId]
	return ok
}
