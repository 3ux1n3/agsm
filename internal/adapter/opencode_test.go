package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/3ux1n3/agsm/internal/session"
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

func TestDeleteSessionRemovesFileDirectly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewOpenCodeAdapter(t.TempDir())
	if err := a.DeleteSession(session.Session{ID: "sess-1", FilePath: path}); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat err=%v", err)
	}
}

func TestDeleteSessionFallsBackToCLIWhenFileMissing(t *testing.T) {
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "delete.log")
	script := "#!/bin/sh\nprintf '%s %s %s\n' \"$1\" \"$2\" \"$3\" > \"$AGSM_DELETE_LOG\"\n"
	cmdPath := filepath.Join(binDir, "opencode")
	if err := os.WriteFile(cmdPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("AGSM_DELETE_LOG", logPath)

	a := NewOpenCodeAdapter(t.TempDir())
	err := a.DeleteSession(session.Session{ID: "sess-2", FilePath: filepath.Join(t.TempDir(), "missing.json")})
	if err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected CLI fallback to run: %v", err)
	}
	if string(data) != "session delete sess-2\n" {
		t.Fatalf("unexpected CLI invocation: %q", string(data))
	}
}

func TestFirstStringAndNumberUseBoundedSearch(t *testing.T) {
	payload := map[string]any{
		"id":   "sess-1",
		"time": map[string]any{"updated": json.Number("1770992982299")},
		"messages": []any{
			map[string]any{
				"tool": map[string]any{
					"result": map[string]any{
						"path": "/tmp/too-deep",
					},
				},
			},
		},
	}

	if got := firstNumber(payload, "updated"); got != 1770992982299 {
		t.Fatalf("expected shallow updated timestamp, got %d", got)
	}
	if got := firstString(payload, "path"); got != "" {
		t.Fatalf("expected deep path lookup to stop before tool results, got %q", got)
	}
	if got := deepProjectDir(payload); got != "/tmp/too-deep" {
		t.Fatalf("expected deepProjectDir to keep deeper absolute path lookup, got %q", got)
	}
}
