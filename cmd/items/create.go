package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/credentials"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

func createItem() error {
	paths, err := path.New("")
	if err != nil {
		return err
	}

	items := persist.New[int64, wowitem.Item](paths.Items + ".new")
	items.SetDirty()

	err = items.Save()
	if err != nil {
		return fmt.Errorf("failed to save items persist: %w", err)
	}

	fmt.Printf("Created items persist: %s\n", paths.Items+".new")

	return nil
}

// createAppearance creates a new appearance persistence and populates it
func createAppearance() error {
	paths, err := path.New("")
	if err != nil {
		return err
	}

	asaIDs := persist.New[int64, bool](paths.Appearances + ".new")

	appearanceSetsIDs, err := wowapi.ItemAppearanceSetsIndexIDs()
	if err != nil {
		return err
	}

	total := len(appearanceSetsIDs)
	count := 1

	for setID, setName := range appearanceSetsIDs {
		fmt.Fprintf(os.Stderr, "Loading appearance set %4d/%4d: %5d  %s\n", count, total, setID, setName)
		count++
		asIDs, err := wowapi.ItemAppearanceSetIDs(setID)
		if err != nil {
			return err
		}
		for _, appearanceID := range asIDs {
			asaIDs.Set(appearanceID, true)
		}
	}

	err = asaIDs.Save()
	if err != nil {

		return fmt.Errorf("failed to save appearances persist: %w", err)
	}

	fmt.Printf("Created appearance set persist: %s\n", paths.Appearances+".new")

	return nil
}

func runCreate(args []string) error {
	flags := flag.NewFlagSet("create", flag.ExitOnError)

	if err := flags.Parse(args); err != nil {
		return err
	}

	if len(args) != 1 {
		usage()
		return fmt.Errorf("must specify a persistence type")
	}
	persistence := args[0]

	paths, err := path.New("")
	if err != nil {
		return err
	}

	clientID, err := credentials.ReadFromKeychain(paths.Secret, "clientID")
	if err != nil {
		return err
	}

	clientSecret, err := credentials.ReadFromKeychain(paths.Secret, "clientSecret")
	if err != nil {
		return err
	}

	err = wowapi.Authenticate(clientID, clientSecret)
	if err != nil {
		return err
	}

	switch persistence {
	case "item":
		return createItem()
	case "appearance":
		return createAppearance()
	default:
		usage()
		return fmt.Errorf("unknown persistence type: %s", persistence)
	}
}
