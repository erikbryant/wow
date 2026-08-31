package output

import (
	"cmp"
	"embed"
	"slices"
	"strings"
	"text/template"

	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowitem"
)

//go:embed PriceCache.tmpl
var embeddedFS embed.FS

type Price struct {
	ItemID int64
	Price  int64
}

type Cosmetic struct {
	ItemID int64
}

type MerchantData struct {
	Prices    []Price
	Cosmetics []Cosmetic
}

// Lua writes item data as Lua source code.
func Lua(wi *wowitem.Persistence) string {
	data := MerchantData{}

	for _, i := range wi.Values() {
		if i.Cosmetic() {
			data.Cosmetics = append(data.Cosmetics, Cosmetic{ItemID: i.ID()})
		}

		spr := i.SellPriceRealizable()
		if spr > common.Coppers(0, 1, 0) {
			// To keep the lua table compact, ignore anything that can't ever be a bargain.
			// The lowest price in the auction house is 1 silver, so skip sell prices <= 1 silver.
			data.Prices = append(data.Prices, Price{ItemID: i.ID(), Price: spr})
		}
	}

	slices.SortFunc(data.Prices, func(a, b Price) int {
		return cmp.Compare(a.ItemID, b.ItemID)
	})

	slices.SortFunc(data.Cosmetics, func(a, b Cosmetic) int {
		return cmp.Compare(a.ItemID, b.ItemID)
	})

	var buf strings.Builder
	err := template.Must(template.ParseFS(embeddedFS, "PriceCache.tmpl")).Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
