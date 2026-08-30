package common

import "fmt"

// MsaValue returns the value at keys within a nested JSON object.
//
// A nil value is returned if:
//   - keys is empty,
//   - msi is nil,
//   - a key does not exist, or
//   - an intermediate value is not a map[string]any.
//
// This is intended for traversing values decoded from JSON using
// encoding/json.
func MsaValue(msi any, keys []string) (any, error) {
	if len(keys) == 0 {
		return msi, nil
	}

	value := msi

	for _, key := range keys {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"cannot access key %q: value has type %T",
				key,
				value,
			)
		}

		value, ok = object[key]
		if !ok {
			return nil, fmt.Errorf("key %q not found", key)
		}
	}

	return value, nil
}

// MsaValued returns the value at keys within a nested JSON object.
//
// If the path does not exist, or traversal cannot continue because an
// intermediate value is not a map[string]any, d is returned.
func MsaValued(msi any, keys []string, d any) any {
	value, err := MsaValue(msi, keys)
	if err != nil {
		return d
	}

	if value == nil {
		return d
	}

	return value
}
