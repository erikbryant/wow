package main

import (
	"flag"
	"fmt"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowitem"
)

// createItemPersist creates a new, empty item persistence and saves it
func createItemPersist(paths *path.Paths) error {
	items := wowitem.NewEmpty(paths.Items)

	err := items.Save()
	if err != nil {
		return fmt.Errorf("failed to save items persist: %w", err)
	}

	fmt.Printf("Saved items persist %s\n", items.Path())

	return nil
}

// createAppearancePersist creates a new, populated appearance persistence and saves it
func createAppearancePersist(paths *path.Paths) error {
	err := authenticate(paths)
	if err != nil {
		return err
	}

	as, err := appearanceset.NewFromWeb(paths.Appearances)
	if err != nil {
		return fmt.Errorf("failed to create appearance set: %w", err)
	}

	err = as.Save()
	if err != nil {

		return fmt.Errorf("failed to save appearances persist: %w", err)
	}

	fmt.Printf("Saved %d IDs to appearance persist %s\n", as.Len(), as.Path())

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
		return createItemPersist(paths)
	case "appearance":
		return createAppearancePersist(paths)
	default:
		usage()
		return fmt.Errorf("unknown persistence type: %s", persistence)
	}
}
