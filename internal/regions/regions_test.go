package regions

import "testing"

func TestRealmToSlug(t *testing.T) {
	tests := []struct {
		realm string
		want  string
	}{
		{"Stormrage", "stormrage"},
		{"Test Realm", "test-realm"},
		{"Drak'thul", "drakthul"},
		{"Aegwynn", "aegwynn"},
		{"Blackwing-Lair", "blackwinglair"},
	}

	for _, tt := range tests {
		t.Run(tt.realm, func(t *testing.T) {
			got := RealmToSlug(tt.realm)
			if got != tt.want {
				t.Errorf("realmToSlug(%q) = %q, want %q",
					tt.realm, got, tt.want)
			}
		})
	}
}
