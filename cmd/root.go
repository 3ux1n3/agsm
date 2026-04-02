package cmd

import (
	"github.com/3ux1n3/agsm/internal/adapter"
	"github.com/3ux1n3/agsm/internal/config"
	"github.com/3ux1n3/agsm/internal/metadata"
	"github.com/3ux1n3/agsm/internal/registry"
	"github.com/3ux1n3/agsm/internal/tui"
)

func Run() error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	metaStore, err := metadata.NewStore("")
	if err != nil {
		return err
	}

	adapters := []adapter.AgentAdapter{}
	if cfg.Agents.OpenCode.Enabled {
		adapters = append(adapters, adapter.NewOpenCodeAdapter(cfg.Agents.OpenCode.SessionPath))
	}

	reg := registry.New(adapters, metaStore, cfg.SortBy, cfg.SortOrder)

	app := tui.NewApp(cfg, reg)
	return app.Run()

}
