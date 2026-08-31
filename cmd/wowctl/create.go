package main

import (
	"fmt"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowapi"
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
	err := wowapi.Init(paths.Secret)
	if err != nil {
		return err
	}

	client, err := wowapi.NewClient()
	if err != nil {
		return err
	}

	as := appearanceset.NewEmpty(paths.Appearances)

	err = as.LoadFromWeb(client)
	if err != nil {
		return fmt.Errorf("failed to load appearance set from web: %w", err)
	}

	err = as.Save()
	if err != nil {

		return fmt.Errorf("failed to save appearances persist: %w", err)
	}

	fmt.Printf("Saved %d IDs to appearance persist %s\n", as.Len(), as.Path())

	return nil
}

func runCreate(args []string, paths *path.Paths) error {
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
