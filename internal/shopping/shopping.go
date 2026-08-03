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
	arbitragePath = "./exports/arbitrageLatest"
	battlePetPath = "./reports/battlePets"
	iLvlPath      = "./reports/arbitrageWithiLvl"
)

var (
	battlePets *battlepet.BattlePet
	mu         sync.Mutex
	toys       *toy.Toy
	userConfig *userconfig.UserConfig
)

// appendFile appends 'contents' to a file
func appendFile(name, contents string) error {
	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0600)
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

// findPetSpellNeeded returns pet spells for sale that I do not own
func findPetSpellNeeded(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for itemID, itemAuctions := range auctions {
		i, ok := wowitem.Get(itemID)
		if !ok {
			continue
		}
		petID, ok := battlePets.PetSpell(i)
		if !ok {
			continue
		}
		if battlePets.Owned(petID) {
			continue
		}

		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout > userConfig.BattlePetPriceUnownedMax {
				continue
			}
			stats := fmt.Sprintf("%s %s %s", battlePets.Name(petID), common.Gold(auc.Buyout), i.Quality())
			bargains = append(bargains, stats)
		}
	}

	return bargains
}

// findPetNeeded returns pets for sale that I do not own
func findPetNeeded(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for _, petAuction := range auctions[battlepet.PetCageItemID] {
		if battlePets.Owned(petAuction.Pet.SpeciesID) {
			continue
		}
		if petAuction.Buyout <= 0 {
			continue
		}
		if petAuction.Buyout > userConfig.BattlePetPriceUnownedMax {
			continue
		}
		bargains = append(bargains, battlePets.Name(petAuction.Pet.SpeciesID))
	}

	// Include any pets available via spells
	spellBargains := findPetSpellNeeded(auctions)
	bargains = append(bargains, spellBargains...)

	return bargains
}

// findPetBargains returns pets that are likely to sell for more than they are listed
func findPetBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for _, petAuction := range auctions[battlepet.PetCageItemID] {
		_, ok := userConfig.SkipPets[petAuction.Pet.SpeciesID]
		if ok {
			continue
		}
		if petAuction.Buyout <= 0 {
			continue
		}
		if petAuction.Pet.QualityID < common.QualityID("Rare") {
			continue
		}
		if petAuction.Pet.Level < 25 {
			continue
		}
		if petAuction.Buyout > userConfig.BattlePetPriceResellMax {
			continue
		}

		bargains = append(bargains, battlePets.Name(petAuction.Pet.SpeciesID))
	}

	return bargains
}

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction, realm string) ([]string, int64, error) {
	arbitrages := []Arbitrage{}
	totalProfit := int64(0)

	for itemID, itemAuctions := range auctions {
		i, ok := wowitem.Get(itemID)
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

			if i.ItemClassName() == "Profession" && !wowitem.Known(i.ID()) {
				// We have not seen this profession tool before. Add iLevels for it in ilevel.go.
				msg := fmt.Sprintf("%d: {}, // %s (%s)  iLvl: %d\n", i.ID(), i.Name(), i.ItemClassName(), i.ItemLevel())
				err := appendFile(iLvlPath, msg)
				if err != nil {
					return nil, 0, err
				}
				fmt.Println(msg)
			}
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

// findBargains returns auctions selling below our desired prices
func findBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for itemID, itemAuctions := range auctions {
		i, ok := wowitem.Get(itemID)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			// Bargains on toys
			if i.Toy() && !toys.Owned(i) && auc.Buyout <= userConfig.ToyPriceMax {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}

			// Bargains on specific items
			maxPrice, ok := userConfig.UsefulGoods[itemID]
			if ok && auc.Buyout <= maxPrice {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}
		}
	}

	return bargains
}

// findAppearanceBargains returns appearances selling at a discount
func findAppearanceBargains(auctions map[int64][]auction.Auction) []string {
	needed := map[string]bool{}

	for itemID, itemAuctions := range auctions {
		i, ok := wowitem.Get(itemID)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			maxPrice := userConfig.AppearancePriceMax
			appearanceSetSuffix := ""
			if transmog.InAppearanceSet(i.Appearances()) {
				maxPrice = userConfig.AppearancePriceInSetMax
				appearanceSetSuffix = "    ---"
			}

			if auc.Buyout > maxPrice {
				continue
			}

			if !transmog.NeedAppearance(i.Appearances()) {
				continue
			}

			needed[i.Name()+appearanceSetSuffix] = true
		}
	}

	bargains := []string{}
	for name := range needed {
		bargains = append(bargains, name)
	}

	return bargains
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

	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", findPetNeeded(auctions), color.New(color.FgMagenta), summarize)
	shoppingList += fmtShoppingList("Pets to Resell", findPetBargains(auctions), color.New(color.FgGreen), summarize)
	shoppingList += fmtShoppingList("Useful Item Bargains", findBargains(auctions), color.New(color.FgRed), summarize)
	shoppingList += fmtShoppingList("Appearance Bargains", findAppearanceBargains(auctions), color.New(color.FgBlue), summarize)

	arbitrages, profit, err := findArbitrages(auctions, realm)
	if err != nil {
		return err
	}

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
func scanRealms(r string, summarize bool) error {
	realms := strings.Split(r, ",")
	results := []string{}
	c := make(chan string)

	// Ensure arbitrage log file is empty
	err := os.WriteFile(arbitragePath, nil, 0600)
	if err != nil {
		return err
	}

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
	fmt.Println(results)

	err = wowitem.Items.Save()
	if err != nil {
		return err
	}

	return nil
}

func Shop(realms string, summarize bool) error {
	var err error

	userConfig = userconfig.New()

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

	// Ensure log file is empty
	err = os.WriteFile(battlePetPath, nil, 0600)
	if err != nil {
		return err
	}
	err = appendFile(battlePetPath, battlePets.Output())
	if err != nil {
		return err
	}

	recipesNeeded, err := cooking.RecipesNeeded()
	if err != nil {
		return err
	}
	for _, recipeName := range recipesNeeded {
		userConfig.UsefulGoods[wowitem.Search(recipeName).ID()] = userConfig.RecipePriceMax
	}

	err = scanRealms(realms, summarize)
	if err != nil {
		return err
	}

	return nil
}
