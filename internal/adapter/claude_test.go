package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/3ux1n3/agsm/internal/session"
)

func TestClaudeDiscover(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-test-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":\"build the release flow\"},\"cwd\":\"/tmp/work\",\"sessionId\":\"12345678-1234-1234-1234-123456789abc\",\"timestamp\":\"2026-04-03T05:13:48Z\"}\n" +
		"{\"type\":\"assistant\",\"slug\":\"steady-heron\",\"cwd\":\"/tmp/work\",\"sessionId\":\"12345678-1234-1234-1234-123456789abc\",\"timestamp\":\"2026-04-03T05:14:48Z\"}\n"
	if err := os.WriteFile(filepath.Join(projectDir, "12345678-1234-1234-1234-123456789abc.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewClaudeAdapter(root)
	items, err := a.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "12345678-1234-1234-1234-123456789abc" {
		t.Fatalf("unexpected id: %#v", items[0])
	}
	if items[0].ProjectDir != "/tmp/work" {
		t.Fatalf("unexpected project dir: %#v", items[0])
	}
	if items[0].Name != "steady-heron" {
		t.Fatalf("unexpected session name: %#v", items[0])
	}
}

func TestClaudeDiscoverIgnoresSubagents(t *testing.T) {
	root := t.TempDir()
	subagents := filepath.Join(root, "proj", "session", "subagents")
	if err := os.MkdirAll(subagents, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subagents, "agent.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewClaudeAdapter(root)
	items, err := a.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no items, got %#v", items)
	}
}

func TestClaudeNewCommandUsesNameAndPrompt(t *testing.T) {
	a := NewClaudeAdapter(t.TempDir())
	cmd := a.NewCommand(NewSessionOptions{Dir: "/tmp/work", Name: "release prep", Prompt: "review this repo"})
	if cmd.Dir != "/tmp/work" {
		t.Fatalf("unexpected dir: %s", cmd.Dir)
	}
	if len(cmd.Args) != 4 || cmd.Args[1] != "--name" || cmd.Args[2] != "release prep" || cmd.Args[3] != "review this repo" {
		t.Fatalf("unexpected args: %#v", cmd.Args)
	}
}

func TestClaudeResumeCommandUsesSessionID(t *testing.T) {
	a := NewClaudeAdapter(t.TempDir())
	cmd := a.ResumeCommand(sessionFixture())
	if len(cmd.Args) != 3 || cmd.Args[1] != "--resume" || cmd.Args[2] != "sess-1" {
		t.Fatalf("unexpected args: %#v", cmd.Args)
	}
}

func TestClaudeDeleteSessionRemovesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewClaudeAdapter(t.TempDir())
	if err := a.DeleteSession(sessionFixtureWithPath(path)); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat err=%v", err)
	}
}

func TestClaudeDiscoverUsesTranscriptTimestampOverModTime(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "session.jsonl")
	content := "{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":\"hello\"},\"cwd\":\"/tmp/work\",\"sessionId\":\"sess-1\",\"timestamp\":\"2026-04-01T10:00:00Z\"}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	modTime := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}

	a := NewClaudeAdapter(root)
	items, err := a.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	want := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	if !items[0].LastActive.Equal(want) {
		t.Fatalf("expected transcript timestamp %v, got %v", want, items[0].LastActive)
	}
}

func TestClaudeDiscoverHandlesLargeJSONLRecord(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "session.jsonl")
	largePrompt := strings.Repeat("a", 3*1024*1024)
	content := fmt.Sprintf("{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":%q},\"cwd\":\"/tmp/work\",\"sessionId\":\"sess-1\",\"timestamp\":\"2026-04-03T05:13:48Z\"}\n", largePrompt)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewClaudeAdapter(root)
	items, err := a.Discover()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "sess-1" {
		t.Fatalf("unexpected session: %#v", items[0])
	}
}

func TestCleanClaudePromptPreservesUTF8(t *testing.T) {
	prompt := strings.Repeat("你", 81)
	got := cleanClaudePrompt(prompt)
	if !utf8.ValidString(got) {
		t.Fatalf("expected valid UTF-8, got %q", got)
	}
	if len([]rune(got)) != 80 {
		t.Fatalf("expected 80 runes, got %d", len([]rune(got)))
	}
}

func sessionFixture() session.Session {
	return sessionFixtureWithPath("/tmp/session.jsonl")
}

func sessionFixtureWithPath(path string) session.Session {
	return session.Session{ID: "sess-1", ProjectDir: "/tmp/work", FilePath: path}
}
