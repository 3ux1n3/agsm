package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const appName = "agsm"

type Config struct {
	SortBy    string `toml:"sort_by"`
	SortOrder string `toml:"sort_order"`
	Agents    Agents `toml:"agents"`
	UI        UI     `toml:"ui"`
}

type Agents struct {
	OpenCode Agent `toml:"opencode"`
}

type Agent struct {
	Enabled     bool   `toml:"enabled"`
	SessionPath string `toml:"session_path"`
}

type UI struct {
	NerdFonts   bool   `toml:"nerd_fonts"`
	ColorScheme string `toml:"color_scheme"`
}

func Default() Config {
	return Config{
		SortBy:    "last_active",
		SortOrder: "desc",
		Agents: Agents{
			OpenCode: Agent{Enabled: true},
		},
		UI: UI{
			NerdFonts:   false,
			ColorScheme: "auto",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		var err error
		path, err = ConfigPath()
		if err != nil {
			return cfg, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	applyDefaults(&cfg)
	return cfg, nil
}

func ConfigDir() (string, error) {
	if override := os.Getenv("AGSM_CONFIG_HOME"); override != "" {
		return override, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func applyDefaults(cfg *Config) {
	defaults := Default()
	cfg.SortBy = strings.ToLower(strings.TrimSpace(cfg.SortBy))
	if cfg.SortBy == "" {
		cfg.SortBy = defaults.SortBy
	}
	switch cfg.SortBy {
	case "last_active", "name", "agent":
	default:
		cfg.SortBy = defaults.SortBy
	}

	cfg.SortOrder = strings.ToLower(strings.TrimSpace(cfg.SortOrder))
	if cfg.SortOrder == "" {
		cfg.SortOrder = defaults.SortOrder
	}
	switch cfg.SortOrder {
	case "asc", "desc":
	default:
		cfg.SortOrder = defaults.SortOrder
	}

	if cfg.UI.ColorScheme == "" {
		cfg.UI.ColorScheme = defaults.UI.ColorScheme
	}
}
