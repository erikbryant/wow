package output

import (
	"encoding/json"
	"io"

	"github.com/erikbryant/wow/internal/wowitem"
)

// JSON writes items as JSON.
func JSON(w io.Writer, items []wowitem.Item) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(items)
}
