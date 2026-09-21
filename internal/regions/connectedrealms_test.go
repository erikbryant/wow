package regions

import (
	"strings"
	"testing"
)

func TestConnectedRealmIDCached(t *testing.T) {
	id, err := ConnectedRealmID("Aegwynn")
	if err != nil {
		t.Fatalf("ConnectedRealmID() error = %v", err)
	}

	if id != connectedRealmIDs["Aegwynn"] {
		t.Errorf("id = %q, want %q",
			id,
			connectedRealmIDs["Aegwynn"])
	}
}

func TestConnectedRealmIDNotFound(t *testing.T) {
	_, err := ConnectedRealmID("Does Not Exist")
	if err == nil {
		t.Fatal("ConnectedRealmID() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "failed to find realm") {
		t.Errorf("error = %q, want realm-not-found error", err)
	}
}
