package adapter

import (
	"os/exec"

	"github.com/3ux1n3/agsm/internal/session"
)

type AgentAdapter interface {
	Name() string
	Discover() ([]session.Session, error)
	ResumeCommand(s session.Session) *exec.Cmd
	NewCommand(dir string) *exec.Cmd
	DeleteSession(s session.Session) error
	IsInstalled() bool
}
