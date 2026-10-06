package commands

import (
	"path/filepath"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/state"
)

// Read operations use the applied runtime names; current settings may describe
// desired values but cannot reconstruct an immutable runtime UUID from a slug.
func recordedProjectForRead(cfg config.ProjectConfig, selector string) (config.ProjectConfig, error) {
	store, err := state.NewStore(cfg.StateDir)
	if err != nil {
		return cfg, err
	}
	var rec state.Record
	if selector != "" {
		rec, _, err = store.StateFileForSelector(selector)
	} else {
		rec, err = store.Load(cfg.Slug)
		if err == nil {
			current, e := filepath.EvalSymlinks(cfg.Dir)
			if e != nil {
				return cfg, e
			}
			recorded, e := filepath.EvalSymlinks(rec.Project.Dir)
			if e != nil {
				return cfg, e
			}
			if current != recorded {
				return cfg, state.ErrOwnerMismatch
			}
		}
	}
	if err != nil {
		return cfg, err
	}
	return rec.Project, nil
}
