package metadata

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/3ux1n3/agsm/internal/config"
)

type Store struct {
	path string
	data fileData
}

type fileData struct {
	CustomNames map[string]string `json:"custom_names"`
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		dir, err := config.EnsureConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(dir, "metadata.json")
	}

	store := &Store{
		path: path,
		data: fileData{CustomNames: map[string]string{}},
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) GetCustomName(key string) string {
	return s.data.CustomNames[key]
}

func (s *Store) SetCustomName(key, value string) error {
	if value == "" {
		delete(s.data.CustomNames, key)
	} else {
		s.data.CustomNames[key] = value
	}
	return s.save()
}

func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.data.CustomNames))
	for key := range s.data.CustomNames {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}
	if s.data.CustomNames == nil {
		s.data.CustomNames = map[string]string{}
	}
	return nil
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0o644)
}
