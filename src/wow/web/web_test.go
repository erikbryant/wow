package web

import (
	"testing"
)

func TestToInt(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected int
	}{
		{int(9), 9},
		{int64(121), 121},
		{string("23"), 23},
		{float64(99.4), 99},
	}

	for _, testCase := range testCases {
		answer := ToInt(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestToInt64(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected int64
	}{
		{int(9), 9},
		{int64(121121121121), 121121121121},
		{string("2323232323"), 2323232323},
		{float64(9988776655.4), 9988776655},
	}

	for _, testCase := range testCases {
		answer := ToInt64(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestToString(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected string
	}{
		{int(9), "9"},
		{int64(121), "121"},
		{string("23"), "23"},
		{float64(99.4), "99.4"},
	}

	for _, testCase := range testCases {
		answer := ToString(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestToFloat64(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected float64
	}{
		{int(9), 9},
		{int64(121), 121},
		{string("23"), 23},
		{float64(99.4), 99.4},
	}

	for _, testCase := range testCases {
		answer := ToFloat64(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}
