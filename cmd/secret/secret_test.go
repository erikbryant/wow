package main

import (
	"strings"
	"testing"
)

func TestRunAddRequiresExactlyOneArgument(t *testing.T) {
	for _, args := range [][]string{nil, {}, {"one", "two"}} {
		if err := runAdd(args); err == nil || !strings.Contains(err.Error(), "add requires exactly 1 argument") {
			t.Errorf("runAdd(%v) error = %v", args, err)
		}
	}
}

func TestRunDeleteRequiresExactlyOneArgument(t *testing.T) {
	for _, args := range [][]string{nil, {}, {"one", "two"}} {
		if err := runDelete(args); err == nil || !strings.Contains(err.Error(), "delete requires exactly one argument") {
			t.Errorf("runDelete(%v) error = %v", args, err)
		}
	}
}

func TestRunGetRequiresAtLeastOneArgument(t *testing.T) {
	if err := runGet(nil); err == nil || !strings.Contains(err.Error(), "get requires at least one argument") {
		t.Fatalf("runGet(nil) error = %v", err)
	}
}
