package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

func createItem(paths *path.Paths) error {
	items := persist.New[int64, wowitem.Item](paths.Items + ".new")

	err := items.Save()
	if err != nil {
		return fmt.Errorf("failed to save items persist: %w", err)
	}

	fmt.Printf("Created items persist: %s\n", paths.Items+".new")

	return nil
}

// createAppearance creates a new appearance persistence and populates it
func createAppearance(paths *path.Paths) error {
	err := authenticate(paths)
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

func runCreate(args []string, paths *path.Paths) error {
	flags := flag.NewFlagSet("create", flag.ExitOnError)

	if err := flags.Parse(args); err != nil {
		return err
	}

	if len(args) != 1 {
		usage()
		return fmt.Errorf("must specify a persistence type")
	}

	persistence := args[0]

	switch persistence {
	case "item":
		return createItem(paths)
	case "appearance":
		return createAppearance(paths)
	default:
		usage()
		return fmt.Errorf("unknown persistence type: %s", persistence)
	}
}
