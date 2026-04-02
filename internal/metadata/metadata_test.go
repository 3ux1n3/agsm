package metadata

import "testing"

func TestStoreSetAndGetCustomName(t *testing.T) {
	store, err := NewStore(t.TempDir() + "/metadata.json")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.SetCustomName("k1", "demo"); err != nil {
		t.Fatal(err)
	}

	if got := store.GetCustomName("k1"); got != "demo" {
		t.Fatalf("unexpected custom name: %s", got)
	}

	if err := store.SetCustomName("k1", ""); err != nil {
		t.Fatal(err)
	}

	if got := store.GetCustomName("k1"); got != "" {
		t.Fatalf("expected empty custom name, got: %s", got)
	}
}
