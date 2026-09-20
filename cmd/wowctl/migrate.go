package main

import (
	"time"

	"github.com/erikbryant/wow/internal/path"
)

type ItemV2 struct {
	// WARNING: Changing this struct invalidates the persistence,
	// even changing the variable names.
	// These members have to be public to write to a gob file,
	// but only use the accessor functions!
	XID      int64
	XItem    map[string]map[string]any // Item JSON, binned by language
	XSource  string
	XUpdated time.Time // Datetime when created or updated
}

func migrate(paths *path.Paths) error {
	//oldItems, err := wowitem.New(paths.Items)
	//if err != nil {
	//	return fmt.Errorf("error loading old items persist: %w", err)
	//}
	//
	//newItems := &struct {
	//	*persist.Persistence[int64, ItemV2]
	//}{
	//	Persistence: persist.New[int64, ItemV2](paths.Items + ".new"),
	//}
	//
	//for _, id := range oldItems.Keys() {
	//	oldItem, ok := oldItems.Persistence.Get(id)
	//	if !ok {
	//		return fmt.Errorf("old items persistence doesn't contain item %q", id)
	//	}
	//	source := oldItem.Source()
	//	if source != "API" {
	//		continue
	//	}
	//	newItem := ItemV2{
	//		XID: oldItem.XID,
	//		XItem: map[string]map[string]any{
	//			"en_US": oldItem.XItem,
	//		},
	//		XSource:  "API",
	//		XUpdated: time.Now(),
	//	}
	//	newItems.Set(oldItem.ID(), newItem)
	//}
	//
	//err = newItems.Save()
	//if err != nil {
	//	return fmt.Errorf("error saving old items persist: %w", err)
	//}
	//
	return nil
}

func runMigrate(args []string, paths *path.Paths) error {
	return migrate(paths)
}
