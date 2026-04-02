package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCodeDiscover(t *testing.T) {
	t.Setenv("PATH", "")

	root := t.TempDir()
	sessionDir := filepath.Join(root, "hash123")
	projectDir := filepath.Join(filepath.Dir(root), "project")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `{"id":"sess-1","title":"debug leak","projectID":"hash123","time":{"updated":1770992982299}}`
	if err := os.WriteFile(filepath.Join(sessionDir, "sess-1.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	projectContent := `{"id":"hash123","worktree":"/tmp/project"}`
	if err := os.WriteFile(filepath.Join(projectDir, "hash123.json"), []byte(projectContent), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewOpenCodeAdapter(root)
	items, err := a.Discover()
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "sess-1" || items[0].ProjectDir != "/tmp/project" {
		t.Fatalf("unexpected session: %#v", items[0])
	}
	if items[0].Name != "debug leak" {
		t.Fatalf("unexpected session name: %#v", items[0])
	}
}
