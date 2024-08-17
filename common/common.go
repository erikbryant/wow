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

// QualityId return the integer id of the given quality name string
func QualityId(quality string) int64 {
	switch strings.ToLower(quality) {
	case "poor":
		return 0
	case "common":
		return 1
	case "uncommon":
		return 2
	case "rare":
		return 3
	case "epic":
		return 4
	case "legendary":
		return 5
	case "artifact":
		return 6
	}

	fmt.Println("ERROR: Unknown quality", quality)
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

// MSIValue returns the value at 'keys' in a map[string]interface{} tree
func MSIValue(msi interface{}, keys []string) (interface{}, error) {
	var ok bool
	value := msi

	for _, key := range keys {
		value, ok = value.(map[string]interface{})[key]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found", key)
		}
	}

	return value, nil
}

// MSIValued returns the value at 'keys' in a map[string]interface{} tree, or a default if value is nil
func MSIValued(msi interface{}, keys []string, d interface{}) (interface{}, error) {
	value, err := MSIValue(msi, keys)
	if value == nil {
		value = d
	}
	return value, err
}
