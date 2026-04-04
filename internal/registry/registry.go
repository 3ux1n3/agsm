package registry

import (
	"sort"
	"strings"
	"sync"

	"github.com/sahilm/fuzzy"

	"github.com/3ux1n3/agsm/internal/adapter"
	"github.com/3ux1n3/agsm/internal/metadata"
	"github.com/3ux1n3/agsm/internal/session"
)

type Registry struct {
	adapters  []adapter.AgentAdapter
	metadata  *metadata.Store
	sortBy    string
	sortOrder string
	mu        sync.RWMutex
	items     []session.Session
}

func New(adapters []adapter.AgentAdapter, metadata *metadata.Store, sortBy, sortOrder string) *Registry {
	return &Registry{
		adapters:  adapters,
		metadata:  metadata,
		sortBy:    sortBy,
		sortOrder: sortOrder,
	}
}

func (r *Registry) Refresh() ([]session.Session, error) {
	all := []session.Session{}
	for _, adapter := range r.adapters {
		items, err := adapter.Discover()
		if err != nil {
			return nil, err
		}
		for i := range items {
			items[i].CustomName = r.metadata.GetCustomName(items[i].MetadataKey())
		}
		all = append(all, items...)
	}

	r.sort(all)
	r.mu.Lock()
	r.items = all
	r.mu.Unlock()
	return r.items, nil
}

func (r *Registry) Items() []session.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]session.Session, len(r.items))
	copy(items, r.items)
	return items
}

func (r *Registry) Rename(s session.Session, name string) error {
	if err := r.metadata.SetCustomName(s.MetadataKey(), strings.TrimSpace(name)); err != nil {
		return err
	}
	r.mu.Lock()
	for i := range r.items {
		if r.items[i].MetadataKey() == s.MetadataKey() {
			r.items[i].CustomName = strings.TrimSpace(name)
		}
	}
	r.mu.Unlock()
	return nil
}

func (r *Registry) Delete(s session.Session) error {
	for _, adapter := range r.adapters {
		if adapter.Name() == s.Agent {
			if err := adapter.DeleteSession(s); err != nil {
				return err
			}
			break
		}
	}
	_, err := r.Refresh()
	return err
}

func (r *Registry) AdapterFor(agent string) adapter.AgentAdapter {
	for _, adapter := range r.adapters {
		if adapter.Name() == agent {
			return adapter
		}
	}
	return nil
}

func (r *Registry) DefaultAdapter() adapter.AgentAdapter {
	if len(r.adapters) == 0 {
		return nil
	}
	return r.adapters[0]
}

func (r *Registry) AdapterCount() int {
	return len(r.adapters)
}

func (r *Registry) AdapterNames() []string {
	names := make([]string, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		names = append(names, adapter.Name())
	}
	return names
}

func (r *Registry) Filter(query string) []session.Session {
	items := r.Items()
	if strings.TrimSpace(query) == "" {
		return items
	}

	targets := make([]string, 0, len(items))
	for _, item := range items {
		targets = append(targets, strings.Join([]string{item.DisplayName(), item.ProjectDir, item.Agent}, " | "))
	}

	matches := fuzzy.Find(strings.TrimSpace(query), targets)
	filtered := make([]session.Session, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, items[match.Index])
	}
	return filtered
}

func (r *Registry) sort(items []session.Session) {
	desc := strings.ToLower(r.sortOrder) != "asc"
	sort.Slice(items, func(i, j int) bool {
		compare := func(left, right session.Session) int {
			switch r.sortBy {
			case "name":
				leftName := strings.ToLower(left.DisplayName())
				rightName := strings.ToLower(right.DisplayName())
				if leftName < rightName {
					return -1
				}
				if leftName > rightName {
					return 1
				}
			case "agent":
				if left.Agent < right.Agent {
					return -1
				}
				if left.Agent > right.Agent {
					return 1
				}
			default:
				if left.LastActive.Before(right.LastActive) {
					return -1
				}
				if left.LastActive.After(right.LastActive) {
					return 1
				}
			}

			if left.Agent < right.Agent {
				return -1
			}
			if left.Agent > right.Agent {
				return 1
			}
			if left.DisplayName() < right.DisplayName() {
				return -1
			}
			if left.DisplayName() > right.DisplayName() {
				return 1
			}
			if left.ID < right.ID {
				return -1
			}
			if left.ID > right.ID {
				return 1
			}
			return 0
		}

		cmp := compare(items[i], items[j])
		switch r.sortBy {
		default:
		}
		if desc {
			return cmp > 0
		}
		return cmp < 0
	})
}
