package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowitem"
)

// Column contains information to retrieve each column of output
type Column struct {
	header string
	value  func(wowitem.Item) string
}

var columns = []Column{
	{
		header: "ID",
		value:  func(item wowitem.Item) string { return fmt.Sprintf("%d", item.ID()) },
	},
	{
		header: "Equips",
		value:  func(item wowitem.Item) string { return fmt.Sprintf("%t", item.Equippable()) },
	},
	{
		header: "Stacks",
		value:  func(item wowitem.Item) string { return fmt.Sprintf("%t", item.Stackable()) },
	},
	{
		header: "App Set",
		value: func(item wowitem.Item) string {
			t := appearanceset.New()
			return fmt.Sprintf("%t", t.Contains(item.Appearances()))
		},
	},
	{
		header: "Sell Price",
		value:  func(item wowitem.Item) string { return common.Gold(item.SellPriceAdvertised()) },
	},
	{
		header: "iLvl",
		value:  func(item wowitem.Item) string { return fmt.Sprintf("%d", item.ItemLevel()) },
	},
	{
		header: "Class",
		value:  func(item wowitem.Item) string { return item.ItemClassName() },
	},
	{
		header: "Quality",
		value:  func(item wowitem.Item) string { return item.Quality() },
	},
	{
		header: "Updated",
		value:  func(item wowitem.Item) string { return item.Updated().Format("2006-01-02") },
	},
	{
		header: "Name",
		value:  func(item wowitem.Item) string { return item.Name() },
	},
}

func headers() (string, string) {
	cols := []string{}
	seps := []string{}

	for _, column := range columns {
		cols = append(cols, column.header)
		seps = append(seps, "-----------------------"[0:len(column.header)])
	}

	return strings.Join(cols, "\t"), strings.Join(seps, "\t")
}

func row(item wowitem.Item) string {
	fields := []string{}

	for _, col := range columns {
		fields = append(fields, col.value(item))
	}

	return strings.Join(fields, "\t")
}

// Table writes items as a human-readable table.
func Table(w io.Writer, items []wowitem.Item) {
	writer := tabwriter.NewWriter(
		w,
		0,
		0,
		2,
		' ',
		0,
	)

	header, separator := headers()

	fmt.Fprintln(writer, header)
	fmt.Fprintln(writer, separator)

	for _, item := range items {
		fmt.Fprintf(writer, "%s\n", row(item))
	}

	writer.Flush()
}
