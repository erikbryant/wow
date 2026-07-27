package output

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/erikbryant/wow/internal/wowitem"
)

// Table writes items as a human-readable table.
func Table(w io.Writer, items []wowitem.Item) error {
	writer := tabwriter.NewWriter(
		w,
		0,
		0,
		2,
		' ',
		0,
	)

	fmt.Fprintln(
		writer,
		"ID\tNAME\tLEVEL\tASet\tQUALITY\tCLASS",
	)

	fmt.Fprintln(
		writer,
		"--\t----\t-----\t----\t-------\t-----",
	)

	for _, item := range items {
		fmt.Fprintf(
			writer,
			"%d\t%s\t%d\t%t\t%s\t%s\n",
			item.ID(),
			item.Name(),
			item.ItemLevel(),
			item.AppearanceSet(),
			item.Quality(),
			item.ItemClassName(),
		)
	}

	return writer.Flush()
}
