package registry

import (
	"os/exec"
	"testing"
	"time"

	"github.com/3ux1n3/agsm/internal/adapter"
	"github.com/3ux1n3/agsm/internal/metadata"
	"github.com/3ux1n3/agsm/internal/session"
)

type fakeAdapter struct {
	items []session.Session
}

func (f fakeAdapter) Name() string                              { return "opencode" }
func (f fakeAdapter) Discover() ([]session.Session, error)      { return f.items, nil }
func (f fakeAdapter) ResumeCommand(s session.Session) *exec.Cmd { return exec.Command("true") }
func (f fakeAdapter) NewCommand(dir string) *exec.Cmd           { return exec.Command("true") }
func (f fakeAdapter) DeleteSession(s session.Session) error     { return nil }
func (f fakeAdapter) IsInstalled() bool                         { return true }

func TestRegistryRefreshSortAndFilter(t *testing.T) {
	meta, err := metadata.NewStore(t.TempDir() + "/metadata.json")
	if err != nil {
		t.Fatal(err)
	}

	items := []session.Session{
		{ID: "1", Agent: "opencode", Name: "older", LastActive: time.Now().Add(-time.Hour), FilePath: "/tmp/1"},
		{ID: "2", Agent: "opencode", Name: "newer", LastActive: time.Now(), FilePath: "/tmp/2"},
	}

	proper := New([]adapter.AgentAdapter{fakeAdapter{items: items}}, meta, "last_active", "desc")
	got, err := proper.Refresh()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0].ID != "2" {
		t.Fatalf("unexpected order: %#v", got)
	}

	filtered := proper.Filter("older")
	if len(filtered) != 1 || filtered[0].ID != "1" {
		t.Fatalf("unexpected filter result: %#v", filtered)
	}
}
