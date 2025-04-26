package transmog

import (
	"fmt"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	allOwned = map[int64]bool{}
)

func Init() {
	allOwned = owned()
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
