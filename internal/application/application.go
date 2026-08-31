package application

import (
	"flag"
	"fmt"
	"strings"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/cooking"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/shoppingconfig"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type App struct {
	Realms []string

	// Initialize these first; some of the others depend on them
	Paths   *path.Paths
	WowItem *wowitem.Persistence

	AppearanceSet  *appearanceset.Persistence
	Appearances    *userconfig.Appearances
	BattlePets     *battlepet.BattlePet
	Cooking        *cooking.CookingRecipes
	ShoppingConfig *shoppingconfig.UserConfig
	Toys           *toy.Toy
	WowAPI         *wowapi.Client
}

// New initializes all singleton data stores
func New(rootPath string) (*App, error) {
	realms := flag.String("realms", userconfig.RealmsWithAlts, "WoW realm(s) to scan")
	flag.Parse()

	var err error
	app := App{
		Realms: strings.Split(*realms, ","),
	}

	for i := range app.Realms {
		app.Realms[i] = strings.TrimSpace(app.Realms[i])
	}

	app.Paths, err = path.New(rootPath)
	if err != nil {
		return nil, err
	}

	err = wowapi.Init(app.Paths.Secret)
	if err != nil {
		return nil, err
	}
	app.WowAPI, err = wowapi.NewClient()
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

	app.Cooking, err = cooking.New(app.WowAPI)
	if err != nil {
		return nil, err
	}

	app.ShoppingConfig = shoppingconfig.New(app.WowItem, app.Cooking)

	app.Toys, err = toy.New(app.WowAPI)
	if err != nil {
		return nil, err
	}

	fmt.Printf("-- #Items persisted        : %d\n", app.WowItem.Len())
	fmt.Printf("-- #Appearances owned      : %d/%d\n", app.Appearances.Len(), app.AppearanceSet.Len())
	fmt.Printf("-- #Battlepet species owned: %d/%d\n", app.BattlePets.LenOwned(), app.BattlePets.LenNames())

	return &app, nil
}
