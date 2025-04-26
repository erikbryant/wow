package transmog

import (
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	allOwned = map[int64]bool{}
)

func Init() {
	allOwned = owned()
}

// Appearances returns a list of all item appearance IDs
func Appearances() ([]int64, bool) {
	ids := []int64{}

	for _, slot := range wowAPI.ItemAppearanceSlotIndex() {
		appearances, ok := wowAPI.ItemAppearanceSlot(slot)
		if !ok {
			return nil, false
		}
		if appearances == nil {
			fmt.Println("Error: Appearances: no appearances for slot:", slot)
		}
		for _, appearance := range appearances {
			id := web.ToInt64(appearance.(map[string]interface{})["id"])
			ids = append(ids, id)
		}
	}
	return ids, true
}

// ItemIdsForAppearance returns a list of item IDs that have the given appearance
func ItemIdsForAppearance(appearanceId int64) ([]int64, bool) {
	ids := []int64{}

	appearance, ok := wowAPI.ItemAppearance(appearanceId)
	if !ok {
		return nil, false
	}

	items := appearance.(map[string]interface{})["items"].([]interface{})

	for _, i := range items {
		id := web.ToInt64(i.(map[string]interface{})["id"])
		ids = append(ids, id)
	}

	return ids, true
}

// owned returns the transmogs I own
func owned() map[int64]bool {
	myTransmogs := map[int64]bool{}

	transmogs, ok := wowAPI.CollectionsTransmogs()
	if !ok {
		log.Fatal("ERROR: Unable to obtain transmogs owned.")
	}

	fmt.Printf("Transmogs keys:")
	for key := range transmogs.(map[string]interface{}) {
		fmt.Printf(" %s", key)
	}
	fmt.Println()

	//for _, transmogRaw := range transmogs {
	//	transmog := transmogRaw.(map[string]interface{})
	//	id, _ := web.MsiValued(transmog, []string{"transmog", "id"}, 0)
	//	myTransmogs[web.ToInt64(id)] = true
	//}

	return myTransmogs
}

//func Own(i item.Item) bool {
//	if len(allOwned) == 0 {
//		log.Fatal("ERROR: You must call transmog.Init() before calling transmog.Own()")
//	}
//
//	for transmogId, name := range allNames {
//		if i.Name() == name {
//			return allOwned[transmogId]
//		}
//	}
//
//	return false
//}
