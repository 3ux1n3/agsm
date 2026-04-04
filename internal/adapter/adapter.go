package adapter

import (
	"os/exec"

	"github.com/3ux1n3/agsm/internal/session"
)

type NewSessionOptions struct {
	Dir    string
	Name   string
	Prompt string
}

type AgentAdapter interface {
	Name() string
	Discover() ([]session.Session, error)
	ResumeCommand(s session.Session) *exec.Cmd
	NewCommand(opts NewSessionOptions) *exec.Cmd
	DeleteSession(s session.Session) error
	IsInstalled() bool
}
