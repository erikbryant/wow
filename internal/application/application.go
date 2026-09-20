package application

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/query"
	"github.com/erikbryant/wow/internal/shoppingconfig"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type App struct {
	Realms userconfig.Realms

	// Initialize these first; some of the others depend on them
	Paths   *path.Paths
	WowItem *wowitem.Persistence

	AppearanceSet  *appearanceset.Persistence
	Appearances    *userconfig.Appearances
	Arbitrages     map[string][]string
	BattlePets     *battlepet.BattlePet
	ShoppingConfig *shoppingconfig.UserConfig
	Toys           *toy.Toy
	WowAPI         *wowapi.Client
}

// New initializes all singleton data stores
func New(rootPath string) (*App, error) {
	var err error
	app := App{
		Arbitrages: make(map[string][]string),
	}

	app.Paths, err = path.New(rootPath)
	if err != nil {
		return nil, err
	}

	app.WowAPI, err = wowapi.NewClient(app.Paths.Secret)
	if err != nil {
		return nil, err
	}

	app.WowItem, err = wowitem.New(app.Paths.Items)
	if err != nil {
		return nil, err
	}

	app.AppearanceSet, err = appearanceset.New(app.Paths.Appearances)
	if err != nil {
		return nil, err
	}

	app.Appearances, err = userconfig.NewAppearances(app.WowAPI)
	if err != nil {
		return nil, err
	}

	app.BattlePets, err = battlepet.New(app.WowAPI)
	if err != nil {
		return nil, err
	}

	app.ShoppingConfig = shoppingconfig.New(app.WowItem)

	app.Toys, err = toy.New(app.WowAPI)
	if err != nil {
		return nil, err
	}

	fmt.Printf("-- #Items persisted        : %d\n", app.WowItem.Len())
	fmt.Printf("-- #Appearances owned      : %d/%d\n", app.Appearances.Len(), app.AppearanceSet.Len())
	fmt.Printf("-- #Battlepet species owned: %d/%d\n", app.BattlePets.LenOwned(), app.BattlePets.LenNames())

	return &app, nil
}

// writeFile creates a new file and writes data into it
func writeFile(path string, data []byte) error {
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func unpack(s string) []string {
	l := strings.Split(s, "\n")
	slices.Sort(l)
	return slices.Compact(l)
}

func (a *App) Shop(shop func(app *App) (string, string, string)) error {

	// ---------------------------- US ----------------------------

	a.Realms = userconfig.RealmsWithAltsUS

	outputBrief, outputVerbose, arbitrageRecords := shop(a)

	// Shopping recommendations
	fmt.Println(outputBrief)

	err := writeFile(a.Paths.RecommendationsBrief, []byte(outputBrief))
	if err != nil {
		return err
	}

	err = writeFile(a.Paths.Recommendations, []byte(outputVerbose))
	if err != nil {
		return err
	}

	// Arbitrages file for the WoW 'wowMerchant' addon to consume
	err = writeFile(a.Paths.Arbitrage, []byte(arbitrageRecords))
	if err != nil {
		return err
	}

	a.Arbitrages[a.Realms.Region] = unpack(arbitrageRecords)

	// ---------------------------- EU ----------------------------

	a.Realms = userconfig.RealmsWithAltsEU

	outputBrief, outputVerbose, arbitrageRecords = shop(a)

	// Shopping recommendations
	fmt.Println(outputVerbose)

	a.Arbitrages[a.Realms.Region] = unpack(arbitrageRecords)

	return nil
}

func (a *App) GenerateOutput() error {
	// Battle pet IDs/names
	err := writeFile(a.Paths.BattlePets, []byte(a.BattlePets.Output()))
	if err != nil {
		return err
	}

	// Arbitrages file for the WoW 'wowMerchant' addon to consume
	err = writeFile(a.Paths.ArbitrageCache, []byte(output.ArbitrageCacheLua(a.Arbitrages)))
	if err != nil {
		return err
	}

	// Prices file for the WoW 'wowMerchant' addon to consume
	err = writeFile(a.Paths.PriceCache, []byte(output.PriceCacheLua(a.WowItem)))
	if err != nil {
		return err
	}

	// Item levels we think we need, but have not encountered yet
	err = writeFile(a.Paths.ILevels, []byte(strings.Join(wowitem.ILevelsNeeded(), "\n")+"\n"))
	if err != nil {
		return err
	}

	// Store persisted items in text form as a backup in case we lose the persistence.
	var buf bytes.Buffer
	items := a.WowItem.Values()
	query.Sort(items, query.ByID)
	output.Table(&buf, items, a.AppearanceSet)
	err = writeFile(a.Paths.ItemsReport, buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
