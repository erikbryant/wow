package shopping

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/erikbryant/wow/internal/application"
	"github.com/erikbryant/wow/internal/auction"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowitem"
)

// Recommendations holds all recommended auctions for a single realm
type Recommendations struct {
	AppearanceBargains []string
	ArbitrageLogs      []string
	ArbitrageProfit    int64
	Arbitrages         []string
	Bargains           []string
	NumUniqueItems     int
	PetNeededBargains  []string
	PetResellBargains  []string
	Realm              string
	Err                error
}

// petSpellNeeded returns true if we do not have this pet and it is a good price
func petSpellNeeded(i wowitem.Item, auc auction.Auction, app *application.App) bool {
	petID, ok := app.BattlePets.PetSpell(i)
	return ok && !app.BattlePets.Owned(petID) && auc.Buyout <= app.ShoppingConfig.BattlePetPriceUnownedMax
}

// petNeeded returns true if we do not have this pet and it is a good price
func petNeeded(petAuction auction.Auction, app *application.App) bool {
	return !app.BattlePets.Owned(petAuction.Pet.SpeciesID) && petAuction.Buyout <= app.ShoppingConfig.BattlePetPriceUnownedMax
}

// petResellBargain returns true if pet is likely to resell at a profit
func petResellBargain(petAuction auction.Auction, app *application.App) bool {
	_, ok := app.ShoppingConfig.SkipPets[petAuction.Pet.SpeciesID]
	if ok {
		return false
	}
	if petAuction.Pet.QualityID < common.QualityID("Rare") {
		return false
	}
	if petAuction.Pet.Level < 25 {
		return false
	}
	if petAuction.Buyout > app.ShoppingConfig.BattlePetPriceResellMax {
		return false
	}
	return true
}

// missingProfessionTool returns true if we do not have an entry for this tool in wowitem/ilevel.go
func missingProfessionTool(i wowitem.Item, app *application.App) bool {
	if i.SellPriceRealizable() <= app.ShoppingConfig.ArbitrageProfitMin {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return false
	}
	return i.ItemClassName() == "Profession" && !wowitem.Known(i.ID())
}

// isArbitrage returns true if the item for auction sells to a vendor for more than the auction price
func isArbitrage(i wowitem.Item, auc auction.Auction, app *application.App) (int64, bool) {
	if auc.Buyout >= i.SellPriceRealizable() {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return 0, false
	}
	profit := (i.SellPriceRealizable() - auc.Buyout) * auc.Quantity
	if profit < app.ShoppingConfig.ArbitrageProfitMin {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return 0, false
	}
	return profit, true
}

// toyBargain returns true if we need this toy, and it is at or below our price
func toyBargain(i wowitem.Item, auc auction.Auction, app *application.App) bool {
	// Bargains on toys
	return i.Toy() && !app.Toys.Owned(i) && auc.Buyout <= app.ShoppingConfig.ToyPriceMax
}

// usefulGoodsBargain returns true if the item for auction is at or below our price
func usefulGoodsBargain(i wowitem.Item, auc auction.Auction, app *application.App) bool {
	maxPrice, ok := app.ShoppingConfig.UsefulGoods[i.ID()]
	return ok && auc.Buyout <= maxPrice
}

// appearanceBargain returns true if the item for auction provides an appearance we need at a good price
func appearanceBargain(i wowitem.Item, auc auction.Auction, app *application.App) bool {
	return auc.Buyout <= app.ShoppingConfig.AppearancePriceMax && app.Appearances.Need(i.Appearances())
}

// appearanceSetBargain returns true if the item for auction provides an appearance (that is in a set) we need at a good price
func appearanceSetBargain(i wowitem.Item, auc auction.Auction, app *application.App) bool {
	return auc.Buyout <= app.ShoppingConfig.AppearancePriceInSetMax && app.AppearanceSet.Contains(i.Appearances()) && app.Appearances.Need(i.Appearances())
}

// iterateAuctions iterates over a single auction house, checking each auction for recommendation
func (r *Recommendations) iterateAuctions(auctions map[int64][]auction.Auction, commodities bool, app *application.App) {
	for itemID, itemAuctions := range auctions {
		i, err := app.WowItem.Get(itemID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting itemID %d, commodities=%t: %v\n", itemID, commodities, err)
			continue
		}

		if missingProfessionTool(i, app) {
			// We have not seen this profession tool yet; add it to wowitem/ilevel.go
			fmt.Fprintf(os.Stderr, "%d: {}, // %s iLvl: %d\n", i.ID(), i.Name(), i.ItemLevel())
		}

		for _, auc := range itemAuctions {

			// ----- Business logic applicable to commodities and regular auctions -----

			profit, ok := isArbitrage(i, auc, app)
			if ok {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(profit))
				r.Arbitrages = append(r.Arbitrages, str)
				r.ArbitrageProfit += profit
				if !commodities {
					for _, iLevel := range wowitem.ILevels(i.ID()) {
						record := fmt.Sprintf("    {%d, %d}, -- %s", i.ID(), iLevel, i.Name())
						r.ArbitrageLogs = append(r.ArbitrageLogs, record)
					}
				}
			}

			if toyBargain(i, auc, app) || usefulGoodsBargain(i, auc, app) {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				r.Bargains = append(r.Bargains, str)
			}

			if commodities {
				continue
			}

			// ----- Business logic applicable only to regular auctions -----

			if i.ID() == battlepet.PetCageItemID {
				if petResellBargain(auc, app) {
					r.PetResellBargains = append(r.PetResellBargains, app.BattlePets.Name(auc.Pet.SpeciesID))
				}
				if petNeeded(auc, app) {
					r.PetNeededBargains = append(r.PetNeededBargains, app.BattlePets.Name(auc.Pet.SpeciesID))
				}
				continue
			}

			if petSpellNeeded(i, auc, app) {
				petID, _ := app.BattlePets.PetSpell(i)
				pet := fmt.Sprintf("%s %s (spell)", app.BattlePets.Name(petID), i.Quality())
				r.PetNeededBargains = append(r.PetNeededBargains, pet)
			}

			if toyBargain(i, auc, app) {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				r.Bargains = append(r.Bargains, str)
			}

			if appearanceSetBargain(i, auc, app) {
				r.AppearanceBargains = append(r.AppearanceBargains, i.Name()+" ---")
			} else {
				// The item is already a bargain, no need to check again
				if appearanceBargain(i, auc, app) {
					r.AppearanceBargains = append(r.AppearanceBargains, i.Name())
				}
			}
		}
	}
}

// scanRealm retrieves auctions and prints suggestions for what to buy for a single realm
func scanRealm(realm string, c chan<- Recommendations, app *application.App) {
	r := Recommendations{
		Realm: realm,
	}

	auctions, err := auction.Get(realm)
	if err != nil {
		r.Err = err
		c <- r
		return
	}

	r.NumUniqueItems = len(auctions)
	r.iterateAuctions(auctions, realm == "Commodities", app)

	c <- r
}

// scanRealms processes auctions on all realms in 'r'
func scanRealms(r string, app *application.App) []Recommendations {
	realms := strings.Split(r, ",")
	results := []Recommendations{}
	c := make(chan Recommendations)

	for _, realm := range realms {
		realm = strings.TrimSpace(realm)
		go scanRealm(realm, c, app)
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

	compact := slices.Clone(items)
	slices.Sort(compact)
	compact = slices.Compact(compact)

	return output.Colorize(fmt.Sprintf("%s%s\n", header, strings.Join(compact, "\n")), fgColor)
}

// format converts a Recommendations to a string
func (r *Recommendations) format(app *application.App, summarize bool) string {
	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", r.PetNeededBargains, output.FgMagenta, summarize)
	shoppingList += fmtShoppingList("Pets to Resell", r.PetResellBargains, output.FgGreen, summarize)
	shoppingList += fmtShoppingList("Useful Item Bargains", r.Bargains, output.FgRed, summarize)
	shoppingList += fmtShoppingList("Appearance Bargains", r.AppearanceBargains, output.FgBlue, summarize)

	if summarize {
		if r.ArbitrageProfit >= app.ShoppingConfig.ProfitToDisplayMin {
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

	msg := output.Colorize(fmt.Sprintf("\n===========>  %s (%d unique items)  <===========\n%s", realm, r.NumUniqueItems, shoppingList), output.FgCyan)

	return msg
}

// generateOutput handles all output (console and files) for shopping
func generateOutput(app *application.App, recommendations []Recommendations) error {
	outputBrief := []string{}
	outputVerbose := []string{}
	arbitrageRecords := []string{}

	for _, r := range recommendations {
		for _, record := range r.ArbitrageLogs {
			arbitrageRecords = append(arbitrageRecords, record)
		}
		outputBrief = append(outputBrief, r.format(app, true))
		outputVerbose = append(outputVerbose, r.format(app, false))
	}

	sort.Strings(outputBrief)
	sort.Strings(outputVerbose)

	fmt.Println(strings.Join(outputBrief, ""))

	// Arbitrages file for the WoW 'wowMerchant' addon to consume
	err := os.WriteFile(app.Paths.Arbitrage, []byte(strings.Join(arbitrageRecords, "\n")+"\n"), 0600)
	if err != nil {
		return err
	}

	// Battle pet IDs/names
	err = os.WriteFile(app.Paths.BattlePets, []byte(app.BattlePets.Output()), 0600)
	if err != nil {
		return err
	}

	// Prices file for the WoW 'wowMerchant' addon to consume
	err = os.WriteFile(app.Paths.PriceCache, []byte(output.Lua(app.WowItem)), 0600)
	if err != nil {
		return err
	}

	// Recipes needed
	err = os.WriteFile(app.Paths.RecipesNeeded, []byte(app.Cooking.Output()), 0600)
	if err != nil {
		return err
	}

	// Verbose form of the shopping recommendations
	err = os.WriteFile(app.Paths.Recommendations, []byte(strings.Join(outputVerbose, "")), 0600)
	if err != nil {
		return err
	}

	return nil
}

// Shop looks for auction house values across the requested realms
func Shop(realms string, app *application.App) error {
	var err error

	recommendations := scanRealms(realms, app)

	err = generateOutput(app, recommendations)
	if err != nil {
		return err
	}

	// Most runs do not change the persistence; be frugal about whether to save
	if app.WowItem.Dirty() {
		err = app.WowItem.Save()
		if err != nil {
			// This is just a cache; failure to save is not fatal
			fmt.Fprintf(os.Stderr, "WARNING: failed to save wow items persistence: %s\n", err)
		}
	}

	return nil
}
