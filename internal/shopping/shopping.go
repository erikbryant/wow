package shopping

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/auction"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/cooking"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/shoppingconfig"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowitem"
)

// Recommendations holds all recommended auctions for a single realm
type Recommendations struct {
	AppearanceBargains []string
	ArbitrageLogs      []string
	ArbitrageProfit    int64
	Arbitrages         []string
	Bargains           []string
	NumAuctions        int
	PetNeededBargains  []string
	PetResellBargains  []string
	Realm              string
	Err                error
}

type DataStore struct {
	// Initialize this first; some of the others depend on it
	WowItem *wowitem.WoWItem

	AppearanceSet    *appearanceset.AppearanceSets
	AppearancesOwned *userconfig.AppearancesOwned
	BattlePets       *battlepet.BattlePet
	CookingRecipes   *cooking.CookingRecipe
	ShoppingConfig   *shoppingconfig.UserConfig
	Toys             *toy.Toy
}

const (
	arbitragePath       = "./exports/arbitrageLatest"
	battlePetPath       = "./reports/battlePets"
	priceCachePath      = "./exports/PriceCache.lua"
	recipesNeededPath   = "./reports/recipesNeeded"
	recommendationsPath = "./reports/shopping"
)

// NewDataStore initializes all singleton data stores
func NewDataStore() (*DataStore, error) {
	var err error
	ds := DataStore{}

	ds.WowItem = wowitem.New()

	ds.AppearanceSet, err = appearanceset.New()
	if err != nil {
		return nil, err
	}

	ds.AppearancesOwned, err = userconfig.NewAppearancesOwned()
	if err != nil {
		return nil, err
	}

	ds.BattlePets, err = battlepet.New()
	if err != nil {
		return nil, err
	}

	ds.CookingRecipes = cooking.New()

	ds.ShoppingConfig = shoppingconfig.New(ds.WowItem, ds.CookingRecipes)

	ds.Toys, err = toy.New()
	if err != nil {
		return nil, err
	}

	return &ds, nil
}

// petSpellNeeded returns true if we do not have this pet and it is a good price
func petSpellNeeded(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	petID, ok := ds.BattlePets.PetSpell(i)
	return ok && !ds.BattlePets.Owned(petID) && auc.Buyout <= ds.ShoppingConfig.BattlePetPriceUnownedMax
}

// petNeeded returns true if we do not have this pet and it is a good price
func petNeeded(petAuction auction.Auction, ds *DataStore) bool {
	return !ds.BattlePets.Owned(petAuction.Pet.SpeciesID) && petAuction.Buyout <= ds.ShoppingConfig.BattlePetPriceUnownedMax
}

// petResellBargain returns true if pet is likely to resell at a profit
func petResellBargain(petAuction auction.Auction, ds *DataStore) bool {
	_, ok := ds.ShoppingConfig.SkipPets[petAuction.Pet.SpeciesID]
	if ok {
		return false
	}
	if petAuction.Pet.QualityID < common.QualityID("Rare") {
		return false
	}
	if petAuction.Pet.Level < 25 {
		return false
	}
	if petAuction.Buyout > ds.ShoppingConfig.BattlePetPriceResellMax {
		return false
	}
	return true
}

// missingProfessionTool returns true if we do not have an entry for this tool in wowitem/ilevel.go
func missingProfessionTool(i wowitem.Item, ds *DataStore) bool {
	if i.SellPriceRealizable() <= ds.ShoppingConfig.ArbitrageProfitMin {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return false
	}
	return i.ItemClassName() == "Profession" && !wowitem.Known(i.ID())
}

// isArbitrage returns true if the item for auction sells to a vendor for more than the auction price
func isArbitrage(i wowitem.Item, auc auction.Auction, ds *DataStore) (int64, bool) {
	if auc.Buyout >= i.SellPriceRealizable() {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return 0, false
	}
	profit := (i.SellPriceRealizable() - auc.Buyout) * auc.Quantity
	if profit < ds.ShoppingConfig.ArbitrageProfitMin {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return 0, false
	}
	return profit, true
}

// toyBargain returns true if we need this toy, and it is at or below our price
func toyBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	// Bargains on toys
	return i.Toy() && !ds.Toys.Owned(i) && auc.Buyout <= ds.ShoppingConfig.ToyPriceMax
}

// usefulGoodsBargain returns true if the item for auction is at or below our price
func usefulGoodsBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	maxPrice, ok := ds.ShoppingConfig.UsefulGoods[i.ID()]
	return ok && auc.Buyout <= maxPrice
}

// appearanceBargain returns true if the item for auction provides an appearance we need at a good price
func appearanceBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	return auc.Buyout <= ds.ShoppingConfig.AppearancePriceMax && ds.AppearancesOwned.Need(i.Appearances())
}

// appearanceSetBargain returns true if the item for auction provides an appearance (that is in a set) we need at a good price
func appearanceSetBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	return auc.Buyout <= ds.ShoppingConfig.AppearancePriceInSetMax && ds.AppearanceSet.Contains(i.Appearances()) && ds.AppearancesOwned.Need(i.Appearances())
}

// iterateAuctions iterates over a single auction house, checking each auction for recommendation
func (r *Recommendations) iterateAuctions(auctions map[int64][]auction.Auction, ds *DataStore) {
	for itemID, itemAuctions := range auctions {
		i, err := ds.WowItem.Get(itemID)
		if err != nil {
			continue
		}

		if missingProfessionTool(i, ds) {
			// We have not seen this profession tool yet; add it to wowitem/ilevel.go
			fmt.Fprintf(os.Stderr, "%d: {}, // %s iLvl: %d\n", i.ID(), i.Name(), i.ItemLevel())
		}

		for _, auc := range itemAuctions {
			if i.ID() == battlepet.PetCageItemID {
				if petResellBargain(auc, ds) {
					r.PetResellBargains = append(r.PetResellBargains, ds.BattlePets.Name(auc.Pet.SpeciesID))
				}
				if petNeeded(auc, ds) {
					r.PetNeededBargains = append(r.PetNeededBargains, ds.BattlePets.Name(auc.Pet.SpeciesID))
				}
				continue
			}

			if petSpellNeeded(i, auc, ds) {
				petID, _ := ds.BattlePets.PetSpell(i)
				pet := fmt.Sprintf("%s %s (spell)", ds.BattlePets.Name(petID), i.Quality())
				r.PetNeededBargains = append(r.PetNeededBargains, pet)
			}

			if toyBargain(i, auc, ds) || usefulGoodsBargain(i, auc, ds) {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				r.Bargains = append(r.Bargains, str)
			}

			if appearanceSetBargain(i, auc, ds) {
				r.AppearanceBargains = append(r.AppearanceBargains, i.Name()+" ---")
			} else {
				// The item is already a bargain, no need to check again
				if appearanceBargain(i, auc, ds) {
					r.AppearanceBargains = append(r.AppearanceBargains, i.Name())
				}
			}

			profit, ok := isArbitrage(i, auc, ds)
			if ok {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(profit))
				r.Arbitrages = append(r.Arbitrages, str)
				r.ArbitrageProfit += profit
				for _, iLevel := range wowitem.ILevels(i.ID()) {
					record := fmt.Sprintf("    {%d, %d}, -- %s", i.ID(), iLevel, i.Name())
					r.ArbitrageLogs = append(r.ArbitrageLogs, record)
				}
			}
		}
	}
}

// scanRealm retrieves auctions and prints suggestions for what to buy for a single realm
func scanRealm(realm string, c chan<- Recommendations, ds *DataStore) {
	r := Recommendations{
		Realm: realm,
	}

	auctions, err := auction.Get(realm)
	if err != nil {
		r.Err = err
		c <- r
		return
	}

	r.NumAuctions = len(auctions)
	r.iterateAuctions(auctions, ds)

	c <- r
}

// scanRealms processes auctions on all realms in 'r'
func scanRealms(r string, ds *DataStore) []Recommendations {
	realms := strings.Split(r, ",")
	results := []Recommendations{}
	c := make(chan Recommendations)

	for _, realm := range realms {
		go scanRealm(realm, c, ds)
	}

	for range len(realms) {
		r := <-c
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "*** failed to scan realm %s: %s\n", r.Realm, r.Err)
			continue
		}
		results = append(results, r)
	}

	return results
}

// fmtShoppingList returns a formatted string of the given items or "" if none
func fmtShoppingList(label string, items []string, fgColor int, summarize bool) string {
	if len(items) == 0 {
		return ""
	}
	header := ""
	if !summarize {
		header = fmt.Sprintf("--- %s ---\n", label)
	}
	slices.Sort(items)
	return output.Colorize(fmt.Sprintf("%s%s\n", header, strings.Join(slices.Compact(items), "\n")), fgColor)
}

// format converts a Recommendations to a string
func (r *Recommendations) format(ds *DataStore, summarize bool) string {
	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", r.PetNeededBargains, output.FgMagenta, summarize)
	shoppingList += fmtShoppingList("Pets to Resell", r.PetResellBargains, output.FgGreen, summarize)
	shoppingList += fmtShoppingList("Useful Item Bargains", r.Bargains, output.FgRed, summarize)
	shoppingList += fmtShoppingList("Appearance Bargains", r.AppearanceBargains, output.FgBlue, summarize)

	if summarize {
		if r.ArbitrageProfit >= ds.ShoppingConfig.ProfitToDisplayMin {
			shoppingList += output.Colorize(fmt.Sprintf("Arbitrages: %s\n", common.Gold(r.ArbitrageProfit)), output.FgWhite)
		}
	} else {
		shoppingList += fmtShoppingList("Arbitrages", r.Arbitrages, output.FgWhite, summarize)
	}

	if len(shoppingList) == 0 {
		// Nothing to buy
		return ""
	}

	// Hack to get Commodities to sort to end of output
	realm := r.Realm
	if realm == "Commodities" {
		realm = "_Commodities_"
	}

	msg := output.Colorize(fmt.Sprintf("\n===========>  %s (%d unique items)  <===========\n%s", realm, r.NumAuctions, shoppingList), output.FgCyan)

	return msg
}

// generateOutput handles all output (console and files) for shopping
func generateOutput(ds *DataStore, recommendations []Recommendations) error {
	outputBrief := []string{}
	outputVerbose := []string{}
	arbitrageRecords := []string{}

	for _, r := range recommendations {
		if r.Realm != "Commodities" {
			for _, record := range r.ArbitrageLogs {
				arbitrageRecords = append(arbitrageRecords, record)
			}
		}
		outputBrief = append(outputBrief, r.format(ds, true))
		outputVerbose = append(outputVerbose, r.format(ds, false))
	}

	sort.Strings(outputBrief)
	sort.Strings(outputVerbose)

	fmt.Println(strings.Join(outputBrief, ""))

	// Arbitrages file for the WoW 'wowMerchant' addon to consume
	err := os.WriteFile(arbitragePath, []byte(strings.Join(arbitrageRecords, "\n")+"\n"), 0600)
	if err != nil {
		return err
	}

	// Battle pet IDs/names
	err = os.WriteFile(battlePetPath, []byte(ds.BattlePets.Output()), 0600)
	if err != nil {
		return err
	}

	// Prices file for the WoW 'wowMerchant' addon to consume
	err = os.WriteFile(priceCachePath, []byte(ds.WowItem.Lua()), 0600)
	if err != nil {
		return err
	}

	// Recipes needed
	err = os.WriteFile(recipesNeededPath, []byte(ds.CookingRecipes.Output()), 0600)
	if err != nil {
		return err
	}

	// Verbose form of the shopping recommendations
	err = os.WriteFile(recommendationsPath, []byte(strings.Join(outputVerbose, "")), 0600)
	if err != nil {
		return err
	}

	return nil
}

// Shop looks for auction house values across all realms
func Shop(realms string) error {
	var err error

	ds, err := NewDataStore()
	if err != nil {
		return err
	}

	recommendations := scanRealms(realms, ds)

	err = ds.WowItem.Items.Save()
	if err != nil {
		return fmt.Errorf("failed to save wow items persistence: %s", err)
	}

	err = generateOutput(ds, recommendations)
	if err != nil {
		return err
	}

	return nil
}
