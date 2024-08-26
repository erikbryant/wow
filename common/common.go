package common

import (
	"fmt"
	"sort"
	"strings"
)

// Coins returns a single numeric value of the given denominations
func Coins(g, s, c int64) int64 {
	return g*100*100 + s*100 + c
}

// Gold returns a formatted string of the given numeric value
func Gold(price int64) string {
	copper := price % 100
	price /= 100
	silver := price % 100
	price /= 100
	gold := price
	return fmt.Sprintf("%d.%02d.%02d", gold, silver, copper)
}

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

// QualityName return the quality name of the given id
func QualityName(qualityId int64) string {
	return qualities[qualityId]
}

// QualityId return the integer id of the given quality name string
func QualityId(qualityName string) int64 {
	for qId, qName := range qualities {
		if strings.ToLower(qName) == strings.ToLower(qualityName) {
			return qId
		}
	}
	fmt.Println("ERROR: Unknown quality", qualityName)
	return -1
}

// SortUnique returns a sorted and unique slice
func SortUnique(values []string) []string {
	alreadySeen := map[string]bool{}
	unique := []string{}

	for _, val := range values {
		if alreadySeen[val] {
			continue
		}
		alreadySeen[val] = true
		unique = append(unique, val)
	}

	sort.Strings(unique)

	return unique
}
