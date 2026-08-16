package wowapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := NewClientWithHTTP(
		"test-client-id",
		"test-client-secret",
		server.URL,
		server.Client(),
	)

	client.accessToken = "test-access-token"
	client.profileAccessToken = "test-profile-access-token"

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("unable to encode JSON: %v", err)
	}
}

func TestClientRequest(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q",
				got, "Bearer test-token")
		}

		writeJSON(t, w, map[string]any{
			"foo": "bar",
		})
	}))

	response, err := client.request(
		client.apiBase+"/test",
		"test-token",
		"TestRequest",
	)
	if err != nil {
		t.Fatalf("request() error = %v", err)
	}

	result, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("response type = %T, want map[string]any", response)
	}

	if result["foo"] != "bar" {
		t.Errorf("foo = %v, want bar", result["foo"])
	}
}

func TestClientRequestHTTPError(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}))

	_, err := client.request(
		client.apiBase+"/test",
		"test-token",
		"TestRequest",
	)
	if err == nil {
		t.Fatal("request() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "HTTP status 500") {
		t.Errorf("error = %q, want HTTP status 500", err)
	}
}

func TestClientRequestInvalidJSON(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"this is not valid JSON`)
	}))

	_, err := client.request(
		client.apiBase+"/test",
		"test-token",
		"TestRequest",
	)
	if err == nil {
		t.Fatal("request() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "unable to decode response") {
		t.Errorf("error = %q, want JSON decoding error", err)
	}
}

func TestRequestKey(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"items": []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
			},
		})
	}))

	result, err := client.requestKey(
		client.apiBase+"/test",
		"test-token",
		"items",
		"TestRequestKey",
	)
	if err != nil {
		t.Fatalf("requestKey() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
}

func TestRequestKeyMissingKey(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"other": []any{},
		})
	}))

	_, err := client.requestKey(
		client.apiBase+"/test",
		"test-token",
		"items",
		"TestRequestKey",
	)
	if err == nil {
		t.Fatal("requestKey() error = nil, want error")
	}

	if !strings.Contains(err.Error(), `response is missing key "items"`) {
		t.Errorf("error = %q, want missing-key error", err)
	}
}

func TestConnectedRealm(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/connected-realm/123" {
			t.Errorf("path = %q, want %q",
				r.URL.Path,
				"/data/wow/connected-realm/123")
		}

		if r.URL.Query().Get("namespace") != "dynamic-us" {
			t.Errorf("namespace = %q, want dynamic-us",
				r.URL.Query().Get("namespace"))
		}

		if r.URL.Query().Get("locale") != "en_US" {
			t.Errorf("locale = %q, want en_US",
				r.URL.Query().Get("locale"))
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("Authorization = %q, want access token", got)
		}

		writeJSON(t, w, map[string]any{
			"id": 123,
			"realms": []any{
				map[string]any{
					"slug": "aegwynn",
				},
			},
		})
	}))

	result, err := client.ConnectedRealm("123")
	if err != nil {
		t.Fatalf("ConnectedRealm() error = %v", err)
	}

	if result["id"].(float64) != 123 {
		t.Errorf("id = %v, want 123", result["id"])
	}
}

func TestConnectedRealmSearch(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/search/connected-realm" {
			t.Errorf("path = %q, want %q",
				r.URL.Path,
				"/data/wow/search/connected-realm")
		}

		if r.URL.Query().Get("namespace") != "dynamic-us" {
			t.Errorf("namespace = %q, want dynamic-us",
				r.URL.Query().Get("namespace"))
		}

		if r.URL.Query().Get("status.type") != "UP" {
			t.Errorf("status.type = %q, want UP",
				r.URL.Query().Get("status.type"))
		}

		writeJSON(t, w, map[string]any{
			"results": []any{},
		})
	}))

	result, err := client.ConnectedRealmSearch()
	if err != nil {
		t.Fatalf("ConnectedRealmSearch() error = %v", err)
	}

	if _, ok := result["results"]; !ok {
		t.Error("response is missing results")
	}
}

func TestConnectedRealmIDCached(t *testing.T) {
	called := false

	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	id, err := client.ConnectedRealmID("Aegwynn")
	if err != nil {
		t.Fatalf("ConnectedRealmID() error = %v", err)
	}

	if id != connectedRealmIDCache["Aegwynn"] {
		t.Errorf("id = %q, want %q",
			id,
			connectedRealmIDCache["Aegwynn"])
	}

	if called {
		t.Error("ConnectedRealmID() made HTTP request for cached realm")
	}
}

func TestConnectedRealmID(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/wow/search/connected-realm":
			writeJSON(t, w, map[string]any{
				"results": []any{
					map[string]any{
						"data": map[string]any{
							"id": 123,
						},
					},
				},
			})

		case "/data/wow/connected-realm/123":
			writeJSON(t, w, map[string]any{
				"realms": []any{
					map[string]any{
						"slug": "test-realm",
					},
				},
			})

		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))

	id, err := client.ConnectedRealmID("Test Realm")
	if err != nil {
		t.Fatalf("ConnectedRealmID() error = %v", err)
	}

	if id != "123" {
		t.Errorf("id = %q, want 123", id)
	}
}

func TestConnectedRealmIDNotFound(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/wow/search/connected-realm":
			writeJSON(t, w, map[string]any{
				"results": []any{},
			})

		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))

	_, err := client.ConnectedRealmID("Does Not Exist")
	if err == nil {
		t.Fatal("ConnectedRealmID() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "failed to find realm") {
		t.Errorf("error = %q, want realm-not-found error", err)
	}
}

func TestAuctions(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/wow/connected-realm/123/auctions":
			if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
				t.Errorf("Authorization = %q, want access token", got)
			}

			writeJSON(t, w, map[string]any{
				"auctions": []any{
					map[string]any{"id": 1},
					map[string]any{"id": 2},
				},
			})

		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))

	// Use a temporary cached realm so Auctions does not have to make
	// additional ConnectedRealm requests.
	const realm = "Test Auctions Realm"
	connectedRealmIDCache[realm] = "123"
	t.Cleanup(func() {
		delete(connectedRealmIDCache, realm)
	})

	auctions, err := client.Auctions(realm)
	if err != nil {
		t.Fatalf("Auctions() error = %v", err)
	}

	if len(auctions) != 2 {
		t.Fatalf("len(auctions) = %d, want 2", len(auctions))
	}
}

func TestCommodities(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/auctions/commodities" {
			t.Errorf("path = %q, want commodities endpoint", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"auctions": []any{
				map[string]any{"item": "ore"},
			},
		})
	}))

	result, err := client.Commodities()
	if err != nil {
		t.Fatalf("Commodities() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
}

func TestItem(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/item/12345" {
			t.Errorf("path = %q, want /data/wow/item/12345", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"id":   12345,
			"name": "Test Item",
		})
	}))

	result, err := client.Item("12345")
	if err != nil {
		t.Fatalf("Item() error = %v", err)
	}

	if result["name"] != "Test Item" {
		t.Errorf("name = %v, want Test Item", result["name"])
	}
}

func TestItemBadStatus(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"status": "nok",
			"reason": "item does not exist",
		})
	}))

	_, err := client.Item("12345")
	if err == nil {
		t.Fatal("Item() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "bad status") {
		t.Errorf("error = %q, want bad-status error", err)
	}
}

func TestPets(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/pet/index" {
			t.Errorf("path = %q, want /data/wow/pet/index", r.URL.Path)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-profile-access-token" {
			t.Errorf("Authorization = %q, want profile token", got)
		}

		writeJSON(t, w, map[string]any{
			"pets": []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
			},
		})
	}))

	result, err := client.Pets()
	if err != nil {
		t.Fatalf("Pets() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
}

func TestCollectionsPets(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profile/user/wow/collections/pets" {
			t.Errorf("path = %q, want collections pets endpoint", r.URL.Path)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-profile-access-token" {
			t.Errorf("Authorization = %q, want profile token", got)
		}

		writeJSON(t, w, map[string]any{
			"pets": []any{
				map[string]any{"id": 1},
			},
		})
	}))

	result, err := client.CollectionsPets()
	if err != nil {
		t.Fatalf("CollectionsPets() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
}

func TestToys(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/toy/index" {
			t.Errorf("path = %q, want /data/wow/toy/index", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"toys": []any{
				map[string]any{"id": 1},
			},
		})
	}))

	result, err := client.Toys()
	if err != nil {
		t.Fatalf("Toys() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
}

func TestCollectionsToys(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profile/user/wow/collections/toys" {
			t.Errorf("path = %q, want collections toys endpoint", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"toys": []any{
				map[string]any{"id": 1},
			},
		})
	}))

	result, err := client.CollectionsToys()
	if err != nil {
		t.Fatalf("CollectionsToys() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
}

func TestItemAppearanceSetsIndex(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/item-appearance/set/index" {
			t.Errorf("path = %q, want appearance set index endpoint", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"appearance_sets": []any{
				map[string]any{
					"id":   100,
					"name": "Test Set",
				},
			},
		})
	}))

	result, err := client.ItemAppearanceSetsIndex()
	if err != nil {
		t.Fatalf("ItemAppearanceSetsIndex() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("len(result) = %d, want 1", len(result))
	}
}

func TestItemAppearanceSetsIndexIDs(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"appearance_sets": []any{
				map[string]any{
					"id":   100,
					"name": "Test Set",
				},
				map[string]any{
					"id":   200,
					"name": "Another Set",
				},
			},
		})
	}))

	result, err := client.ItemAppearanceSetsIndexIDs()
	if err != nil {
		t.Fatalf("ItemAppearanceSetsIndexIDs() error = %v", err)
	}

	if result[100] != "Test Set" {
		t.Errorf("result[100] = %q, want Test Set", result[100])
	}

	if result[200] != "Another Set" {
		t.Errorf("result[200] = %q, want Another Set", result[200])
	}
}

func TestItemAppearanceSet(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/wow/item-appearance/set/123" {
			t.Errorf("path = %q, want appearance set endpoint", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"appearances": []any{
				map[string]any{"id": 10},
				map[string]any{"id": 20},
			},
		})
	}))

	result, err := client.ItemAppearanceSet(123)
	if err != nil {
		t.Fatalf("ItemAppearanceSet() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
}

func TestItemAppearanceSetIDs(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"appearances": []any{
				map[string]any{"id": 10},
				map[string]any{"id": 20},
				map[string]any{"id": 30},
			},
		})
	}))

	result, err := client.ItemAppearanceSetIDs(123)
	if err != nil {
		t.Fatalf("ItemAppearanceSetIDs() error = %v", err)
	}

	want := []int64{10, 20, 30}

	if len(result) != len(want) {
		t.Fatalf("result = %v, want %v", result, want)
	}

	for i := range want {
		if result[i] != want[i] {
			t.Errorf("result[%d] = %d, want %d", i, result[i], want[i])
		}
	}
}

func TestCollectionsTransmogs(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profile/user/wow/collections/transmogs" {
			t.Errorf("path = %q, want transmogs endpoint", r.URL.Path)
		}

		writeJSON(t, w, map[string]any{
			"some": "data",
		})
	}))

	result, err := client.CollectionsTransmogs()
	if err != nil {
		t.Fatalf("CollectionsTransmogs() error = %v", err)
	}

	response, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T, want map[string]any", result)
	}

	if response["some"] != "data" {
		t.Errorf("some = %v, want data", response["some"])
	}
}

func TestProfessions(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profile/wow/character/aegwynn/testcharacter/professions" {
			t.Errorf("path = %q, want professions endpoint", r.URL.Path)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-profile-access-token" {
			t.Errorf("Authorization = %q, want profile token", got)
		}

		writeJSON(t, w, map[string]any{
			"professions": []any{},
		})
	}))

	result, err := client.Professions("Aegwynn", "TestCharacter")
	if err != nil {
		t.Fatalf("Professions() error = %v", err)
	}

	response, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T, want map[string]any", result)
	}

	if _, ok := response["professions"]; !ok {
		t.Error("response is missing professions")
	}
}

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
			got := realmToSlug(tt.realm)
			if got != tt.want {
				t.Errorf("realmToSlug(%q) = %q, want %q",
					tt.realm, got, tt.want)
			}
		})
	}
}

func TestJSONString(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "string",
			value: "123",
			want:  "123",
		},
		{
			name:  "float",
			value: float64(123),
			want:  "123",
		},
		{
			name:  "json number",
			value: json.Number("456"),
			want:  "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonString(tt.value)
			if err != nil {
				t.Fatalf("jsonString() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("jsonString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONInt64(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{
			name:  "float",
			value: float64(123),
			want:  123,
		},
		{
			name:  "int",
			value: int(456),
			want:  456,
		},
		{
			name:  "int64",
			value: int64(789),
			want:  789,
		},
		{
			name:  "json number",
			value: json.Number("123456789"),
			want:  123456789,
		},
		{
			name:  "string",
			value: "987654321",
			want:  987654321,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonInt64(tt.value)
			if err != nil {
				t.Fatalf("jsonInt64() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("jsonInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}
