package common

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	qualities := map[string]int64{
		"poor":      0,
		"common":    1,
		"uncommon":  2,
		"rare":      3,
		"epic":      4,
		"legendary": 5,
		"artifact":  6,
	}

	qID, ok := qualities[strings.ToLower(qualityName)]
	if !ok {
		fmt.Fprintf(os.Stderr, "*** unknown quality: %s\n", qualityName)
		return -1
	}

	return qID
}

// JSONString converts a JSON-decoded value to a string.
func JSONString(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case json.Number:
		return value.String()
	case bool:
		return strconv.FormatBool(value)
	case nil:
		return ""
	default:
		panic(fmt.Sprint(value))
	}
}

// JSONInt64 converts a JSON-decoded value to int64 or returns an error.
func JSONInt64(value any) (int64, error) {
	n, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("expected JSON number, got %T", value)
	}

	return n.Int64()
}

// JSONInt64Panic converts a JSON-decoded value to int64 or panics.
func JSONInt64Panic(value any) int64 {
	n, err := JSONInt64(value)
	if err != nil {
		panic(err)
	}

	return n
}
