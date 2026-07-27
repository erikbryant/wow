package output

import (
	"fmt"
	"io"
	"strings"
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

// Detail writes a single item in a human-readable form.
func Detail(w io.Writer, item wowitem.Item) error {
	fields := []struct {
		Name  string
		Value string
	}{
		{"ID", fmt.Sprintf("%d", item.ID())},
		{"Name", item.Name()},
		{"Level", fmt.Sprintf("%d", item.ItemLevel())},
		{"ASet", fmt.Sprintf("%t", item.AppearanceSet())},
		{"Quality", item.Quality()},
		{"Class", item.ItemClassName()},
	}

	width := 0
	for _, field := range fields {
		if len(field.Name) > width {
			width = len(field.Name)
		}
	}

	for _, field := range fields {
		fmt.Fprintf(
			w,
			"%-*s  %s\n",
			width,
			field.Name,
			field.Value,
		)
	}

	return nil
}

// Summary returns a single-line representation of an item.
// Useful for logs or embedding in other output.
func Summary(item wowitem.Item) string {
	values := []string{
		fmt.Sprintf("%d", item.ID()),
		item.Name(),
		fmt.Sprintf("%d", item.ItemLevel()),
		item.Quality(),
		item.ItemClassName(),
	}

	return strings.Join(values, "\t")
}
