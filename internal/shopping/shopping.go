package shopping

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/auction"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/shoppingconfig"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowitem"
	"github.com/fatih/color"
)

type Recommendations struct {
	Arbitrages         []string
	ArbitrageProfit    int64
	PetNeededBargains  []string
	Bargains           []string
	PetResellBargains  []string
	AppearanceBargains []string
}

type DataStore struct {
	// Initialize this first; some of the others depend on it
	WowItem *wowitem.WoWItem

	BattlePets     *battlepet.BattlePet
	ShoppingConfig *shoppingconfig.UserConfig
	Toys           *toy.Toy

	AppearancesOwned *userconfig.AppearancesOwned
	AppearanceSet    *appearanceset.AppearanceSets
}

const (
	arbitragePath  = "./exports/arbitrageLatest"
	battlePetPath  = "./reports/battlePets"
	priceCachePath = "./exports/PriceCache.lua"
)

var (
	mu               sync.Mutex
	arbitrageRecords []string
)

func NewDataStore() (*DataStore, error) {
	var err error
	ds := DataStore{}

	ds.WowItem = wowitem.New()

	ds.BattlePets, err = battlepet.New()
	if err != nil {
		return nil, err
	}

	ds.ShoppingConfig = shoppingconfig.New(ds.WowItem)

	ds.Toys, err = toy.New()
	if err != nil {
		return nil, err
	}

	ds.AppearancesOwned = userconfig.NewAppearancesOwned()
	ds.AppearanceSet = appearanceset.New()

	return &ds, nil
}

func appendArbitrageRecord(record string) {
	mu.Lock()
	defer mu.Unlock()
	arbitrageRecords = append(arbitrageRecords, record)
}

func petSpellNeeded(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	petID, ok := ds.BattlePets.PetSpell(i)
	return ok && !ds.BattlePets.Owned(petID) && auc.Buyout <= ds.ShoppingConfig.BattlePetPriceUnownedMax
}

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

func missingProfessionTool(i wowitem.Item, ds *DataStore) bool {
	if i.SellPriceRealizable() <= ds.ShoppingConfig.ArbitrageProfitMin {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return false
	}
	return i.ItemClassName() == "Profession" && !wowitem.Known(i.ID())
}

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

// usefulGoodsBargain returns true if it is at or below our price
func usefulGoodsBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	maxPrice, ok := ds.ShoppingConfig.UsefulGoods[i.ID()]
	return ok && auc.Buyout <= maxPrice
}

func appearanceBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	return auc.Buyout <= ds.ShoppingConfig.AppearancePriceMax && ds.AppearancesOwned.Need(i.Appearances())
}

func appearanceSetBargain(i wowitem.Item, auc auction.Auction, ds *DataStore) bool {
	return auc.Buyout <= ds.ShoppingConfig.AppearancePriceInSetMax && ds.AppearanceSet.Has(i.Appearances()) && ds.AppearancesOwned.Need(i.Appearances())
}

// fmtShoppingList returns a formatted string of the given items or "" if none
func fmtShoppingList(label string, items []string, c *color.Color, summarize bool) string {
	if len(items) == 0 {
		return ""
	}
	header := ""
	if !summarize {
		header = fmt.Sprintf("--- %s ---\n", label)
	}
	slices.Sort(items)
	return c.Sprintf("%s%s\n", header, strings.Join(slices.Compact(items), "\n"))
}

func iterateAuctions(auctions map[int64][]auction.Auction, ds *DataStore, saveRecords bool) *Recommendations {
	var recommendations Recommendations

	for itemID, itemAuctions := range auctions {
		i, ok := ds.WowItem.Get(itemID)
		if !ok {
			continue
		}

		if missingProfessionTool(i, ds) {
			// We have not seen this profession tool yet; add it to wowitem/ilevel.go
			fmt.Printf("%d: {}, // %s iLvl: %d\n", i.ID(), i.Name(), i.ItemLevel())
		}

		for _, auc := range itemAuctions {
			if i.ID() == battlepet.PetCageItemID {
				if petResellBargain(auc, ds) {
					recommendations.PetResellBargains = append(recommendations.PetResellBargains, ds.BattlePets.Name(auc.Pet.SpeciesID))
				}
				if petNeeded(auc, ds) {
					recommendations.PetNeededBargains = append(recommendations.PetNeededBargains, ds.BattlePets.Name(auc.Pet.SpeciesID))
				}
				continue
			}

			if petSpellNeeded(i, auc, ds) {
				petID, _ := ds.BattlePets.PetSpell(i)
				pet := fmt.Sprintf("%s %s (spell)", ds.BattlePets.Name(petID), i.Quality())
				recommendations.PetNeededBargains = append(recommendations.PetNeededBargains, pet)
			}

			if toyBargain(i, auc, ds) || usefulGoodsBargain(i, auc, ds) {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				recommendations.Bargains = append(recommendations.Bargains, str)
			}

			if appearanceSetBargain(i, auc, ds) {
				recommendations.AppearanceBargains = append(recommendations.AppearanceBargains, i.Name()+" ---")
			} else {
				// The item is already a bargain, no need to check again
				if appearanceBargain(i, auc, ds) {
					recommendations.AppearanceBargains = append(recommendations.AppearanceBargains, i.Name())
				}
			}

			profit, ok := isArbitrage(i, auc, ds)
			if ok {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(profit))
				recommendations.Arbitrages = append(recommendations.Arbitrages, str)
				recommendations.ArbitrageProfit += profit
				if saveRecords {
					for _, iLevel := range wowitem.ILevels(i.ID()) {
						record := fmt.Sprintf("    {%d, %d}, -- %s", i.ID(), iLevel, i.Name())
						appendArbitrageRecord(record)
					}
				}
			}
		}
	}

	return &recommendations
}

func formatRecommendations(recommendations *Recommendations, realm string, numAuctions int, ds *DataStore, summarize bool) string {
	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", recommendations.PetNeededBargains, color.New(color.FgMagenta), summarize)
	shoppingList += fmtShoppingList("Pets to Resell", recommendations.PetResellBargains, color.New(color.FgGreen), summarize)
	shoppingList += fmtShoppingList("Useful Item Bargains", recommendations.Bargains, color.New(color.FgRed), summarize)
	shoppingList += fmtShoppingList("Appearance Bargains", recommendations.AppearanceBargains, color.New(color.FgBlue), summarize)

	if summarize {
		if recommendations.ArbitrageProfit >= ds.ShoppingConfig.ProfitToDisplayMin {
			c := color.New(color.FgWhite)
			shoppingList += c.Sprintf("Arbitrages: %s\n", common.Gold(recommendations.ArbitrageProfit))
		}
	} else {
		shoppingList += fmtShoppingList("Arbitrages", recommendations.Arbitrages, color.New(color.FgWhite), summarize)
	}

	if len(shoppingList) == 0 {
		// Nothing to buy
		return ""
	}

	col := color.New(color.FgCyan)
	msg := col.Sprintf("\n===========>  %s (%d unique items)  <===========\n%s", realm, numAuctions, shoppingList)

	return msg
}

// scanRealm retrieves auctions and prints suggestions for what to buy for a single realm
func scanRealm(realm string, c chan<- string, ds *DataStore, summarize bool) {
	auctions, err := auction.Get(realm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to scan realm %s: %s\n", realm, err)
		c <- ""
		return
	}

	recommendations := iterateAuctions(auctions, ds, realm != "Commodities")

	c <- formatRecommendations(recommendations, realm, len(auctions), ds, summarize)
}

// scanRealms processes auctions on all realms in 'r'
func scanRealms(r string, ds *DataStore, summarize bool) []string {
	realms := strings.Split(r, ",")
	results := []string{}
	c := make(chan string)

	for _, realm := range realms {
		go scanRealm(realm, c, ds, summarize)
	}

	for range len(realms) {
		s := <-c
		if s == "" {
			continue
		}
		// Hack to get Commodities to sort to end of slice
		s = strings.Replace(s, " Commodities ", " _Commodities_ ", 1)
		results = append(results, s)
	}

	sort.Strings(results)

	return results
}

func writeLogs(ds *DataStore) error {
	// Write the battle pet IDs/names
	err := os.WriteFile(battlePetPath, []byte(ds.BattlePets.Output()), 0600)
	if err != nil {
		return err
	}

	// Write the prices file for the WoW 'wowMerchant' addon to consume
	err = os.WriteFile(priceCachePath, []byte(ds.WowItem.Lua()), 0600)
	if err != nil {
		return err
	}

	// Write the arbitrages file for the WoW 'wowMerchant' addon to consume
	mu.Lock()
	err = os.WriteFile(arbitragePath, []byte(strings.Join(arbitrageRecords, "\n")+"\n"), 0600)
	mu.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func Shop(realms string, summarize bool) error {
	var err error

	ds, err := NewDataStore()
	defer func() {
		err := ds.WowItem.Items.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to save wow items persistence: %s\n", err)
		}
	}()

	results := scanRealms(realms, ds, summarize)
	fmt.Println(results)

	err = writeLogs(ds)
	if err != nil {
		return err
	}

	return nil
}
