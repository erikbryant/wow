package shopping

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/erikbryant/wow/internal/auction"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/cooking"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/transmog"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowitem"
	"github.com/fatih/color"
)

type Arbitrage struct {
	item   wowitem.Item
	profit int64
}

const (
	arbitragePath  = "./exports/arbitrageLatest"
	battlePetPath  = "./reports/battlePets"
	priceCachePath = "./exports/PriceCache.lua"
)

var (
	battlePets *battlepet.BattlePet
	mu         sync.Mutex
	toys       *toy.Toy
	userConfig *userconfig.UserConfig
	wowItem    *wowitem.WoWItem
)

// appendFile appends 'contents' to a file
func appendFile(name, contents string) error {
	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %s", name, err)
	}
	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		return fmt.Errorf("failed to write to log file %s: %s", name, err)
	}

	return nil
}

func petSpellNeeded(i wowitem.Item, auc auction.Auction) bool {
	petID, ok := battlePets.PetSpell(i)
	return ok && !battlePets.Owned(petID) && auc.Buyout <= userConfig.BattlePetPriceUnownedMax
}

func petNeeded(petAuction auction.Auction) bool {
	return !battlePets.Owned(petAuction.Pet.SpeciesID) && petAuction.Buyout <= userConfig.BattlePetPriceUnownedMax
}

// petResellBargain returns true if pet is likely to resell at a profit
func petResellBargain(petAuction auction.Auction) bool {
	_, ok := userConfig.SkipPets[petAuction.Pet.SpeciesID]
	if ok {
		return false
	}
	if petAuction.Buyout <= 0 {
		return false
	}
	if petAuction.Pet.QualityID < common.QualityID("Rare") {
		return false
	}
	if petAuction.Pet.Level < 25 {
		return false
	}
	if petAuction.Buyout > userConfig.BattlePetPriceResellMax {
		return false
	}
	return true
}

func missingProfessionTool(i wowitem.Item) bool {
	if i.SellPriceRealizable() <= userConfig.ArbitrageProfitMin {
		// Not enough profit to make it worth the WoW runtime it takes to scan the AH
		return false
	}
	return i.ItemClassName() == "Profession" && !wowitem.Known(i.ID())
}

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction, realm string) ([]string, int64, error) {
	arbitrages := []Arbitrage{}
	totalProfit := int64(0)

	for itemID, itemAuctions := range auctions {
		i, ok := wowItem.Get(itemID)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout >= i.SellPriceRealizable() {
				continue
			}
			profit := (i.SellPriceRealizable() - auc.Buyout) * auc.Quantity
			if profit < userConfig.ArbitrageProfitMin {
				// Not enough profit to make it worth the WoW runtime it takes to scan the AH
				continue
			}

			arbitrages = append(arbitrages, Arbitrage{i, profit})
		}
	}

	bargains := []string{}
	for _, arbitrage := range arbitrages {
		totalProfit += arbitrage.profit

		str := fmt.Sprintf("%s   %s", arbitrage.item.Name(), common.Gold(arbitrage.profit))
		bargains = append(bargains, str)

		if realm == "Commodities" {
			// Commodities are not worth logging; their prices are too volatile
			continue
		}

		iLevels := wowitem.ILevels(arbitrage.item.ID())
		for _, iLevel := range iLevels {
			logEntry := fmt.Sprintf("    {%d, %d}, -- %s\n", arbitrage.item.ID(), iLevel, arbitrage.item.Name())
			err := appendFile(arbitragePath, logEntry)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	slices.Sort(bargains)
	bargains = slices.Compact(bargains)

	return bargains, totalProfit, nil
}

// toyBargain returns true if we need this toy, and it is at or below our price
func toyBargain(i wowitem.Item, auc auction.Auction) bool {
	// Bargains on toys
	return i.Toy() && !toys.Owned(i) && auc.Buyout <= userConfig.ToyPriceMax
}

// usefulGoodsBargain returns true if it is at or below our price
func usefulGoodsBargain(i wowitem.Item, auc auction.Auction) bool {
	maxPrice, ok := userConfig.UsefulGoods[i.ID()]
	return ok && auc.Buyout <= maxPrice
}

func appearanceBargain(i wowitem.Item, auc auction.Auction) bool {
	return auc.Buyout <= userConfig.AppearancePriceMax && transmog.NeedAppearance(i.Appearances())
}

func appearanceSetBargain(i wowitem.Item, auc auction.Auction) bool {
	return auc.Buyout <= userConfig.AppearancePriceInSetMax && transmog.InAppearanceSet(i.Appearances()) && transmog.NeedAppearance(i.Appearances())
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
	return c.Sprintf("%s%s\n", header, strings.Join(common.SortUnique(items), "\n"))
}

// scanRealm retrieves auctions and prints suggestions for what to buy for a single realm
func scanRealm(realm string, c chan<- string, summarize bool) error {
	auctions, err := auction.Get(realm)
	if err != nil {
		c <- ""
		return err
	}

	arbitrages, profit, err := findArbitrages(auctions, realm)
	if err != nil {
		return err
	}

	petNeededBargains := []string{}
	bargains := []string{}
	petResellBargains := []string{}
	appearanceBargains := []string{}

	for itemID, itemAuctions := range auctions {
		i, ok := wowItem.Get(itemID)
		if !ok {
			continue
		}

		if missingProfessionTool(i) {
			// We have not seen this profession tool before. Add iLevels for it in ilevel.go.
			fmt.Printf("%d: {}, // %s iLvl: %d\n", i.ID(), i.Name(), i.ItemLevel())
		}

		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			if i.ID() == battlepet.PetCageItemID {
				if petResellBargain(auc) {
					petResellBargains = append(petResellBargains, battlePets.Name(auc.Pet.SpeciesID))
				}
				if petNeeded(auc) {
					petNeededBargains = append(petResellBargains, battlePets.Name(auc.Pet.SpeciesID))
				}
				continue
			}

			if petSpellNeeded(i, auc) {
				petID, _ := battlePets.PetSpell(i)
				pet := fmt.Sprintf("%s %s (spell)", battlePets.Name(petID), i.Quality())
				petNeededBargains = append(petResellBargains, pet)
			}

			if toyBargain(i, auc) || usefulGoodsBargain(i, auc) {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}

			if appearanceSetBargain(i, auc) {
				appearanceBargains = append(appearanceBargains, i.Name()+" ---")
			} else {
				// If the item is already a bargain, no need to check again
				if appearanceBargain(i, auc) {
					appearanceBargains = append(appearanceBargains, i.Name())
				}
			}
		}
	}

	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", petNeededBargains, color.New(color.FgMagenta), summarize)
	shoppingList += fmtShoppingList("Pets to Resell", petResellBargains, color.New(color.FgGreen), summarize)
	shoppingList += fmtShoppingList("Useful Item Bargains", bargains, color.New(color.FgRed), summarize)
	shoppingList += fmtShoppingList("Appearance Bargains", appearanceBargains, color.New(color.FgBlue), summarize)

	if summarize {
		if profit >= userConfig.ProfitToDisplayMin {
			c := color.New(color.FgWhite)
			shoppingList += c.Sprintf("Arbitrages: %s\n", common.Gold(profit))
		}
	} else {
		shoppingList += fmtShoppingList("Arbitrages", arbitrages, color.New(color.FgWhite), summarize)
	}

	if len(shoppingList) == 0 {
		// Nothing to buy
		c <- ""
		return nil
	}

	col := color.New(color.FgCyan)
	c <- col.Sprintf("\n===========>  %s (%d unique items)  <===========\n%s", realm, len(auctions), shoppingList)

	return nil
}

// scanRealms processes auctions on all realms in 'r'
func scanRealms(r string, summarize bool) ([]string, error) {
	realms := strings.Split(r, ",")
	results := []string{}
	c := make(chan string)

	for _, realm := range realms {
		go func() {
			err := scanRealm(realm, c, summarize)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to scan realm %s: %s\n", realm, err)
				os.Exit(1)
			}
		}()
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

	return results, nil
}

func writeLogs() error {
	// Write the battle pet IDs/names
	err := os.WriteFile(battlePetPath, []byte(battlePets.Output()), 0600)
	if err != nil {
		return err
	}

	// Write the prices file for the WoW 'wowMerchant' addon to consume
	err = os.WriteFile(priceCachePath, []byte(wowItem.Lua()), 0600)
	if err != nil {
		return err
	}

	return nil
}

func Shop(realms string, summarize bool) error {
	var err error

	wowItem = wowitem.New()
	defer func() {
		err := wowItem.Items.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to save wow items persistence: %s\n", err)
		}
	}()

	userConfig = userconfig.New(wowItem)

	battlePets, err = battlepet.New()
	if err != nil {
		return err
	}

	toys, err = toy.New()
	if err != nil {
		return err
	}

	err = transmog.Init()
	if err != nil {
		return err
	}

	// Ensure arbitrage log file is empty
	err = os.WriteFile(arbitragePath, nil, 0600)
	if err != nil {
		return err
	}

	recipesNeeded, err := cooking.RecipesNeeded()
	if err != nil {
		return err
	}
	for _, recipeName := range recipesNeeded {
		userConfig.UsefulGoods[wowItem.Search(recipeName).ID()] = userConfig.RecipePriceMax
	}

	results, err := scanRealms(realms, summarize)
	if err != nil {
		return err
	}
	fmt.Println(results)

	err = writeLogs()
	if err != nil {
		return err
	}

	return nil
}
