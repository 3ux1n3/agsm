package adapter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/3ux1n3/agsm/internal/session"
)

type ClaudeAdapter struct {
	sessionPath string
	initErr     error
}

func NewClaudeAdapter(sessionPath string) *ClaudeAdapter {
	if sessionPath != "" {
		return &ClaudeAdapter{sessionPath: sessionPath}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return &ClaudeAdapter{initErr: fmt.Errorf("resolve home directory: %w", err)}
	}
	if home == "" {
		return &ClaudeAdapter{initErr: fmt.Errorf("resolve home directory: empty path")}
	}

	return &ClaudeAdapter{sessionPath: filepath.Join(home, ".claude", "projects")}
}

func (a *ClaudeAdapter) Name() string {
	return "claude"
}

func (a *ClaudeAdapter) Discover() ([]session.Session, error) {
	if a.initErr != nil {
		return nil, a.initErr
	}

	items := []session.Session{}
	seen := map[string]struct{}{}
	err := filepath.WalkDir(a.sessionPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "subagents" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".jsonl" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		s, err := a.parseSession(path, info.ModTime())
		if err != nil || s.ID == "" {
			return nil
		}
		if _, ok := seen[s.ID]; ok {
			return nil
		}
		items = append(items, s)
		seen[s.ID] = struct{}{}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].LastActive.After(items[j].LastActive)
	})

	return items, nil
}

func (a *ClaudeAdapter) ResumeCommand(s session.Session) *exec.Cmd {
	cmd := exec.Command("claude", "--resume", s.ID)
	if s.ProjectDir != "" {
		cmd.Dir = s.ProjectDir
	}
	return cmd
}

func (a *ClaudeAdapter) NewCommand(opts NewSessionOptions) *exec.Cmd {
	args := make([]string, 0, 3)
	if name := strings.TrimSpace(opts.Name); name != "" {
		args = append(args, "--name", name)
	}
	if prompt := strings.TrimSpace(opts.Prompt); prompt != "" {
		args = append(args, prompt)
	}
	cmd := exec.Command("claude", args...)
	cmd.Dir = opts.Dir
	return cmd
}

func (a *ClaudeAdapter) DeleteSession(s session.Session) error {
	if s.FilePath == "" {
		return fmt.Errorf("claude session file path is required")
	}
	if err := os.Remove(s.FilePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (a *ClaudeAdapter) IsInstalled() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

type claudeRecord struct {
	Type      string         `json:"type"`
	SessionID string         `json:"sessionId"`
	Cwd       string         `json:"cwd"`
	Timestamp string         `json:"timestamp"`
	Slug      string         `json:"slug"`
	Message   *claudeMessage `json:"message"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func (a *ClaudeAdapter) parseSession(path string, modTime time.Time) (session.Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return session.Session{}, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	id := ""
	projectDir := ""
	name := ""
	lastActive := time.Time{}
	seenTimestamp := false

	for {
		line, err := readJSONLRecord(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return session.Session{}, err
		}

		var record claudeRecord
		if err := json.Unmarshal(line, &record); err != nil {
			continue
		}
		if record.SessionID != "" {
			id = record.SessionID
		}
		if record.Cwd != "" {
			projectDir = record.Cwd
		}
		if ts, ok := parseTime(record.Timestamp); ok {
			if !seenTimestamp || ts.After(lastActive) {
				lastActive = ts
			}
			seenTimestamp = true
		}
		if slug := strings.TrimSpace(record.Slug); slug != "" {
			name = slug
		}
		if name == "" && record.Message != nil && record.Message.Role == "user" {
			if prompt := firstClaudePrompt(record.Message.Content); prompt != "" {
				name = prompt
			}
		}
	}
	if !seenTimestamp {
		lastActive = modTime
	}

	if id == "" {
		id = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if name == "" {
		projectName := filepath.Base(projectDir)
		if projectName != "." && projectName != string(filepath.Separator) && projectName != "" {
			name = fmt.Sprintf("%s %s", projectName, shortSessionID(id))
		} else {
			name = shortSessionID(id)
		}
	}

	return session.Session{
		ID:         id,
		Agent:      a.Name(),
		Name:       name,
		ProjectDir: projectDir,
		LastActive: lastActive,
		FilePath:   path,
	}, nil
}

func firstClaudePrompt(content any) string {
	switch value := content.(type) {
	case string:
		return cleanClaudePrompt(value)
	case []any:
		for _, item := range value {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if part["type"] != "text" {
				continue
			}
			if text, ok := part["text"].(string); ok {
				if prompt := cleanClaudePrompt(text); prompt != "" {
					return prompt
				}
			}
		}
	}
	return ""
}

func cleanClaudePrompt(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "." {
		return ""
	}
	v = strings.Join(strings.Fields(v), " ")
	if len([]rune(v)) > 80 {
		v = strings.TrimSpace(string([]rune(v)[:80]))
	}
	return v
}

func readJSONLRecord(r *bufio.Reader) ([]byte, error) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					return nil, io.EOF
				}
			} else {
				return nil, err
			}
		}

		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			if err == io.EOF {
				return nil, io.EOF
			}
			continue
		}
		out := make([]byte, len(trimmed))
		copy(out, trimmed)
		return out, nil
	}
}

func shortSessionID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
