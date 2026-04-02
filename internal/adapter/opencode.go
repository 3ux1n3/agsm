package adapter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/3ux1n3/agsm/internal/session"
)

type OpenCodeAdapter struct {
	sessionPath string
	projectPath string
}

func NewOpenCodeAdapter(sessionPath string) *OpenCodeAdapter {
	if sessionPath == "" {
		home, _ := os.UserHomeDir()
		sessionPath = filepath.Join(home, ".local", "share", "opencode", "storage", "session")
		return &OpenCodeAdapter{sessionPath: sessionPath, projectPath: filepath.Join(home, ".local", "share", "opencode", "storage", "project")}
	}
	return &OpenCodeAdapter{sessionPath: sessionPath, projectPath: filepath.Join(filepath.Dir(sessionPath), "project")}
}

func (a *OpenCodeAdapter) Name() string {
	return "opencode"
}

func (a *OpenCodeAdapter) Discover() ([]session.Session, error) {
	items := []session.Session{}
	seen := map[string]struct{}{}
	projectDirs, _ := a.loadProjectDirs()
	dbItems, err := a.discoverFromDB()
	if err == nil {
		for _, item := range dbItems {
			items = append(items, item)
			seen[item.ID] = struct{}{}
		}
	}
	err = filepath.WalkDir(a.sessionPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		s, err := a.parseSession(path, info.ModTime(), projectDirs)
		if err != nil {
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

type openCodeDBSession struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Directory   string `json:"directory"`
	Title       string `json:"title"`
	TimeUpdated int64  `json:"time_updated"`
}

func (a *OpenCodeAdapter) discoverFromDB() ([]session.Session, error) {
	if !a.IsInstalled() {
		return nil, fmt.Errorf("opencode not installed")
	}

	query := "select id, project_id, directory, title, time_updated from session where time_archived is null order by time_updated desc;"
	cmd := exec.Command("opencode", "db", query, "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var rows []openCodeDBSession
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, err
	}

	items := make([]session.Session, 0, len(rows))
	for _, row := range rows {
		if row.ID == "" {
			continue
		}
		name := strings.TrimSpace(row.Title)
		if name == "" {
			name = row.ID
		}
		items = append(items, session.Session{
			ID:         row.ID,
			Agent:      a.Name(),
			Name:       name,
			ProjectDir: row.Directory,
			LastActive: parseUnixMillis(row.TimeUpdated),
			FilePath:   filepath.Join(a.sessionPath, row.ProjectID, row.ID+".json"),
		})
	}

	return items, nil
}

func (a *OpenCodeAdapter) ResumeCommand(s session.Session) *exec.Cmd {
	cmd := exec.Command("opencode", "-s", s.ID)
	if s.ProjectDir != "" {
		cmd.Dir = s.ProjectDir
	}
	return cmd
}

func (a *OpenCodeAdapter) NewCommand(dir string) *exec.Cmd {
	cmd := exec.Command("opencode")
	cmd.Dir = dir
	return cmd
}

func (a *OpenCodeAdapter) DeleteSession(s session.Session) error {
	if s.FilePath != "" {
		if err := os.Remove(s.FilePath); err == nil || os.IsNotExist(err) == false {
			if err == nil {
				return nil
			}
			return err
		}
	}
	cmd := exec.Command("opencode", "session", "delete", s.ID)
	return cmd.Run()
}

func (a *OpenCodeAdapter) IsInstalled() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

func (a *OpenCodeAdapter) parseSession(path string, modTime time.Time, projectDirs map[string]string) (session.Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return session.Session{}, err
	}

	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return session.Session{}, err
	}

	id := firstString(payload, "id", "sessionID", "sessionId")
	if id == "" {
		id = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	name := firstString(payload, "name", "title", "sessionName")
	projectDir := firstString(payload, "projectDir", "projectDirectory", "cwd", "path", "workspace", "directory")
	projectID := firstString(payload, "projectID", "projectId")
	if projectDir == "" && projectID != "" {
		projectDir = projectDirs[projectID]
	}
	if !filepath.IsAbs(projectDir) {
		if candidate := deepProjectDir(payload); candidate != "" {
			projectDir = candidate
		}
	}

	lastActive := modTime
	if ts := firstString(payload, "updatedAt", "lastActive", "timestamp", "createdAt"); ts != "" {
		if parsed, ok := parseTime(ts); ok {
			lastActive = parsed
		}
	} else if ts := firstNumber(payload, "updated", "created", "timestamp"); ts > 0 {
		lastActive = parseUnixMillis(ts)
	}

	if name == "" {
		projectName := filepath.Base(projectDir)
		if projectName != "." && projectName != string(filepath.Separator) && projectName != "" {
			name = fmt.Sprintf("%s %s", projectName, id)
		} else {
			name = id
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

func (a *OpenCodeAdapter) loadProjectDirs() (map[string]string, error) {
	projects := map[string]string{}
	entries, err := os.ReadDir(a.projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return projects, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(a.projectPath, entry.Name()))
		if err != nil {
			continue
		}
		var payload any
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}
		id := firstString(payload, "id")
		worktree := firstString(payload, "worktree", "directory")
		if id != "" && worktree != "" {
			projects[id] = worktree
		}
	}
	return projects, nil
}

func firstString(v any, keys ...string) string {
	switch value := v.(type) {
	case map[string]any:
		for _, key := range keys {
			if raw, ok := value[key]; ok {
				if s, ok := raw.(string); ok && s != "" {
					return s
				}
			}
		}
		for _, child := range value {
			if s := firstString(child, keys...); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range value {
			if s := firstString(child, keys...); s != "" {
				return s
			}
		}
	}
	return ""
}

func firstNumber(v any, keys ...string) int64 {
	switch value := v.(type) {
	case map[string]any:
		for _, key := range keys {
			if raw, ok := value[key]; ok {
				switch n := raw.(type) {
				case float64:
					return int64(n)
				case int64:
					return n
				case json.Number:
					if parsed, err := n.Int64(); err == nil {
						return parsed
					}
				case string:
					if parsed, err := strconv.ParseInt(n, 10, 64); err == nil {
						return parsed
					}
				}
			}
		}
		for _, child := range value {
			if n := firstNumber(child, keys...); n > 0 {
				return n
			}
		}
	case []any:
		for _, child := range value {
			if n := firstNumber(child, keys...); n > 0 {
				return n
			}
		}
	}
	return 0
}

func deepProjectDir(v any) string {
	candidate := firstString(v, "projectDir", "projectDirectory", "cwd", "path", "workspace", "directory")
	if filepath.IsAbs(candidate) {
		return candidate
	}
	return ""
}

func parseTime(v string) (time.Time, bool) {
	formats := []string{time.RFC3339, time.RFC3339Nano, time.DateTime}
	for _, format := range formats {
		if t, err := time.Parse(format, v); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseUnixMillis(v int64) time.Time {
	return time.UnixMilli(v)
}
