package item

var itemLevels = map[int64][]int64{
	191235: {71, 72, 73, 74},     // Draconium Blacksmith's Toolbox (Profession - 317)
	191236: {71, 72, 73, 74},     // Draconium Leatherworker's Toolset (Profession - 317)
	191237: {70, 71, 72, 73, 74}, // Draconium Blacksmith's Hammer (Profession - 317)
	191238: {71, 72, 73, 74},     // Draconium Leatherworker's Knife (Profession - 317)
	191239: {71, 72, 73, 74},     // Draconium Needle Set (Profession - 317)
	191240: {71, 72, 73, 74},     // Draconium Skinning Knife (Profession - 317)

	238018: {212, 218, 232}, // Sun-Blessed Blacksmith's Hammer (Profession - 106)

	193486: {71}, // Resilient Smock (Profession - 317)

	240953: {180, 186, 192, 199, 206}, // Bold Biographer's Bifocals (Profession - 80)
	240954: {180, 186, 192, 199, 206}, // Fantastic Font Focuser (Profession - 80)

	215120: {79, 85, 91, 98, 105},     // Radiant Loupes (Profession - 486)
	240955: {180, 183, 186, 189, 193}, // Silvermoon Loupes (Profession - 80)
	240959: {212, 218, 225, 232},      // Sin'dorei Jeweler's Loupes (Profession - 106)
}

// Known returns true if the item has an entry in itemLevels
func Known(itemId int64) bool {
	_, ok := itemLevels[itemId]
	return ok
}

// ItemLevels returns the item levels seen for this item, or zero if none
func ItemLevels(itemId int64) []int64 {
	iLvls, ok := itemLevels[itemId]
	if !ok {
		return []int64{0}
	}
	return iLvls
}
