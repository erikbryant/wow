package wowitem

import (
	"fmt"
	"slices"
	"strings"
)

// itemIDs returns the sorted list of keys from the persistence file
func itemIDs() []int64 {
	keys := Items.Keys()
	slices.Sort(keys)
	return keys
}

func luaVendorPrice() (string, []string) {
	var lua strings.Builder

	lua.WriteString(fmt.Sprintf("local VendorSellPriceCache = {\n"))
	for _, id := range itemIDs() {
		i, _ := Items.Get(id)
		spr := i.SellPriceRealizable()
		if spr <= 100 {
			// To keep the lua table small, ignore anything that can't ever be a bargain
			// Skip prices that are zero
			// Skip prices <= one silver (the auction house does not deal in copper)
			continue
		}
		lua.WriteString(fmt.Sprintf("  [\"%d\"] = %d,\n", id, spr))
	}
	lua.WriteString(fmt.Sprintf("}\n"))

	lua.WriteString(fmt.Sprintf(`
-- VendorSellPrice returns the cached vendor sell price
local function VendorSellPrice(itemID)
    return VendorSellPriceCache[tostring(itemID)] or 0
end

-- ValidatePriceCache verifies each cached sell price matches the actual sell price
local function ValidatePriceCache()
    for itemID, cachedPrice in pairs(VendorSellPriceCache) do
        itemID = tonumber(itemID)
        local item = Item:CreateFromItemID(itemID)
        item:ContinueOnItemLoad(
                function()
                    local itemInfo = { C_Item.GetItemInfo(itemID) }
                    local sellPrice = itemInfo[11]
                    if cachedPrice ~= sellPrice then
                        MerchUtil.PrettyPrint("Cached price mismatch!", itemID, GetCoinTextureString(cachedPrice), "~=", GetCoinTextureString(sellPrice))
                    end
                end
        )
    end
end
`))

	return lua.String(), []string{"VendorSellPrice", "ValidatePriceCache"}
}

func luaCosmetic() (string, []string) {
	var lua strings.Builder

	lua.WriteString(fmt.Sprintf("local Cosmetics = {\n"))
	for _, id := range itemIDs() {
		i, _ := Items.Get(id)
		cosmetic := i.Cosmetic()
		if !cosmetic {
			continue
		}
		lua.WriteString(fmt.Sprintf("  [\"%d\"] = true,\n", id))
	}
	lua.WriteString(fmt.Sprintf("}\n"))

	lua.WriteString(fmt.Sprintf(`
-- Cosmetic returns true if the item is a Cosmetic
local function Cosmetic(itemID)
    return Cosmetics[tostring(itemID)] or false
end
`))

	return lua.String(), []string{"Cosmetic"}
}

// Lua the cached vendor sell prices to stdout as a lua table and accessor
func Lua() string {
	lua := ""

	lua, fcns := luaVendorPrice()

	lua2, fcns2 := luaCosmetic()

	lua += "\n" + lua2
	for _, fn := range fcns2 {
		fcns = append(fcns, fn)
	}

	lua += fmt.Sprintf("\nPriceCache = {\n")
	for _, fn := range fcns {
		lua += fmt.Sprintf("  %s = %s,\n", fn, fn)
	}
	lua += fmt.Sprintf("}\n")

	return lua
}
