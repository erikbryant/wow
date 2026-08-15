package application

import (
	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/cooking"
	"github.com/erikbryant/wow/internal/credentials"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/shoppingconfig"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type App struct {
	// Initialize this first; some of the others depend on it
	WowItem *wowitem.WoWItem

	AppearanceSet  *appearanceset.AppearanceSets
	Appearances    *userconfig.Appearances
	BattlePets     *battlepet.BattlePet
	CookingRecipes *cooking.CookingRecipe
	Paths          *path.Paths
	ShoppingConfig *shoppingconfig.UserConfig
	Toys           *toy.Toy
}

// TODO: This is a duplicate of authenticate() in cmd/items/authenticate.go

// authenticate authenticates this session against the WoW web APIs
func (app *App) authenticate() error {
	clientID, err := credentials.ReadFromKeychain(app.Paths.Secret, "clientID")
	if err != nil {
		return err
	}

	clientSecret, err := credentials.ReadFromKeychain(app.Paths.Secret, "clientSecret")
	if err != nil {
		return err
	}

	err = wowapi.Authenticate(clientID, clientSecret)
	if err != nil {
		return err
	}

	return nil
}

// New initializes all singleton data stores
func New(rootPath string) (*App, error) {
	var err error
	app := App{}

	app.Paths, err = path.New(rootPath)
	if err != nil {
		return nil, err
	}

	err = app.authenticate()
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

	app.Appearances, err = userconfig.NewAppearances()
	if err != nil {
		return nil, err
	}

	app.BattlePets, err = battlepet.New()
	if err != nil {
		return nil, err
	}

	app.CookingRecipes, err = cooking.New()
	if err != nil {
		return nil, err
	}

	app.ShoppingConfig = shoppingconfig.New(app.WowItem, app.CookingRecipes)

	app.Toys, err = toy.New()
	if err != nil {
		return nil, err
	}

	return &app, nil
}
