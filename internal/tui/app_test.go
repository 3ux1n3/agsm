package tui

import (
	"os/exec"
	"testing"
	"time"

	"github.com/3ux1n3/agsm/internal/adapter"
	"github.com/3ux1n3/agsm/internal/config"
	"github.com/3ux1n3/agsm/internal/metadata"
	"github.com/3ux1n3/agsm/internal/registry"
	"github.com/3ux1n3/agsm/internal/session"
)

type fakeAdapter struct {
	name string
}

func (f fakeAdapter) Name() string                              { return f.name }
func (f fakeAdapter) Discover() ([]session.Session, error)      { return nil, nil }
func (f fakeAdapter) ResumeCommand(s session.Session) *exec.Cmd { return exec.Command("true") }
func (f fakeAdapter) NewCommand(opts adapter.NewSessionOptions) *exec.Cmd {
	return exec.Command("true")
}
func (f fakeAdapter) DeleteSession(s session.Session) error { return nil }
func (f fakeAdapter) IsInstalled() bool                     { return true }

func TestNewSessionDefaultsToCurrentAgent(t *testing.T) {
	app := testApp(t)
	app.items = []session.Session{{ID: "sess-1", Agent: "claude", Name: "demo", LastActive: time.Now()}}
	app.selected = 0

	if got := app.defaultNewAgent(); got != "claude" {
		t.Fatalf("expected current agent to be selected, got %q", got)
	}
}

func TestCycleNewAgentChangesModalAgent(t *testing.T) {
	app := testApp(t)
	app.newAgent = "opencode"

	app.cycleNewAgent(1)
	if app.newAgent != "claude" {
		t.Fatalf("expected next agent to be claude, got %q", app.newAgent)
	}

	app.cycleNewAgent(1)
	if app.newAgent != "opencode" {
		t.Fatalf("expected wraparound to opencode, got %q", app.newAgent)
	}
}

func TestNewSessionFieldCountDependsOnSelectedAgent(t *testing.T) {
	app := testApp(t)
	app.newAgent = "opencode"
	if got := app.newSessionFieldCount(); got != 2 {
		t.Fatalf("expected 2 fields for opencode, got %d", got)
	}

	app.newAgent = "claude"
	if got := app.newSessionFieldCount(); got != 3 {
		t.Fatalf("expected 3 fields for claude, got %d", got)
	}
}

func testApp(t *testing.T) *app {
	t.Helper()
	meta, err := metadata.NewStore(t.TempDir() + "/metadata.json")
	if err != nil {
		t.Fatal(err)
	}
	reg := registry.New([]adapter.AgentAdapter{
		fakeAdapter{name: "opencode"},
		fakeAdapter{name: "claude"},
	}, meta, "last_active", "desc")
	return NewApp(config.Default(), reg)
}
