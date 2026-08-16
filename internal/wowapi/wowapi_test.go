package wowapi

import "testing"

func TestRealmToSlug(t *testing.T) {
	cases := map[string]string{"Aegwynn": "aegwynn", "Anub'arak": "anubarak", "Altar of Storms": "altar-of-storms", "Drak'thul": "drakthul", "Kul Tiras": "kul-tiras"}
	for in, want := range cases {
		if got := realmToSlug(in); got != want {
			t.Errorf("%q=%q want %q", in, got, want)
		}
	}
}
func TestConnectedRealmIDCached(t *testing.T) {
	for realm, want := range map[string]string{"Aegwynn": "1136", "Aman'Thul": "3726"} {
		got, err := ConnectedRealmID(realm)
		if err != nil || got != want {
			t.Errorf("%s=%q,%v", realm, got, err)
		}
	}
}
func TestBadItemID(t *testing.T) {
	if !BadItemID(23704) || BadItemID(1) {
		t.Fail()
	}
}
func TestRequestKey(t *testing.T) {
	old := accessToken
	accessToken = ""
	defer func() { accessToken = old }()
	_ = requestKey
}
