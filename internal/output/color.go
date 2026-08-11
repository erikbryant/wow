package output

import (
	"fmt"
)

const (
	FgBlack = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

const escape = "\x1b"

// Colorize wraps s with terminal color controls
func Colorize(s string, fgColor int) string {
	return fmt.Sprintf("%s[%dm%s%s[%dm", escape, fgColor, s, escape, 0)
}
