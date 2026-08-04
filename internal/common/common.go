package common

import (
	"fmt"
	"os"
	"strings"
)

var (
	qualities = map[int64]string{
		0: "Poor",
		1: "Common",
		2: "Uncommon",
		3: "Rare",
		4: "Epic",
		5: "Legendary",
		6: "Artifact",
	}
)

// Coppers returns the value in coppers of the given denominations
func Coppers(g, s, c int64) int64 {
	return g*100*100 + s*100 + c
}

// Gold returns a formatted string of the given numeric value
func Gold(coppers int64) string {
	c := coppers % 100
	coppers /= 100
	s := coppers % 100
	coppers /= 100
	g := coppers
	return fmt.Sprintf("%d.%02d.%02d", g, s, c)
}

// QualityID return the integer id of the given quality name string
func QualityID(qualityName string) int64 {
	for qID, qName := range qualities {
		if strings.ToLower(qName) == strings.ToLower(qualityName) {
			return qID
		}
	}
	fmt.Fprintf(os.Stderr, "*** unknown quality: %s\n", qualityName)
	return -1
}
