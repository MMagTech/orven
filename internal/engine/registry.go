package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"orven/contract"
)

// Plugin is an installed plugin as the engine sees it: manifest plus
// resolved runtime limits.
type Plugin struct {
	Manifest contract.Manifest
	Dir      string

	Recommended time.Duration
	MinInterval time.Duration
	MaxInterval time.Duration
	Freshness   time.Duration
	Timeout     time.Duration

	Incompatible bool   // requires a newer engine contract
	LoadError    string // manifest problems, shown in plugin list
}

func parseDur(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

// LoadPlugins discovers plugins in dir. Each plugin is a folder with a
// plugin.yaml. Product rule (CONSTRAINTS.md, plugin identity): an
// installation may contain only one plugin with a given plugin ID —
// the ID keys config, secrets, observations, and history, so a second
// folder claiming a loaded ID is marked broken rather than silently
// sharing the first one's storage.
func LoadPlugins(dir string) []*Plugin {
	var out []*Plugin
	seen := map[string]string{} // plugin id -> folder that claimed it
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := LoadPlugin(filepath.Join(dir, e.Name()))
		if p == nil {
			continue
		}
		if p.LoadError == "" {
			if prev, dup := seen[p.Manifest.ID]; dup {
				p.LoadError = fmt.Sprintf(
					"duplicate plugin id %q (already provided by %s) — an installation may contain only one plugin with a given id",
					p.Manifest.ID, prev)
			} else {
				seen[p.Manifest.ID] = e.Name()
			}
		}
		out = append(out, p)
	}
	return out
}

// LoadPlugin loads one plugin folder; nil if the folder has no
// plugin.yaml. A broken manifest yields a Plugin with LoadError set so
// callers can explain why it cannot run, rather than silently hiding
// it. Also used by `orven validate`.
func LoadPlugin(pdir string) *Plugin {
	b, err := os.ReadFile(filepath.Join(pdir, "plugin.yaml"))
	if err != nil {
		return nil // not a plugin folder
	}
	base := filepath.Base(pdir)
	p := &Plugin{Dir: pdir}
	if err := yaml.Unmarshal(b, &p.Manifest); err != nil {
		p.Manifest.ID = base
		p.Manifest.Name = base
		p.LoadError = fmt.Sprintf("invalid plugin.yaml: %v", err)
		return p
	}
	if p.Manifest.ID == "" || len(p.Manifest.Entrypoint) == 0 {
		p.Manifest.ID = base
		if p.Manifest.Name == "" {
			p.Manifest.Name = base
		}
		p.LoadError = "plugin.yaml must declare id and entrypoint"
		return p
	}
	if p.Manifest.Engine.MinContract > contract.Version {
		p.Incompatible = true
	}
	p.Recommended = parseDur(p.Manifest.Collection.RecommendedInterval, 30*time.Minute)
	p.MinInterval = parseDur(p.Manifest.Collection.MinInterval, 5*time.Minute)
	p.MaxInterval = parseDur(p.Manifest.Collection.MaxInterval, 24*time.Hour)
	p.Freshness = parseDur(p.Manifest.Collection.Freshness, 2*p.Recommended)
	p.Timeout = parseDur(p.Manifest.Timeout, 60*time.Second)
	if p.Timeout > 5*time.Minute {
		p.Timeout = 5 * time.Minute // engine-enforced ceiling
	}
	return p
}

// Interval returns the effective collection interval for a plugin,
// honoring the user override but clamping to the plugin's declared
// min/max — the engine, not the plugin or the user, has the last word.
func (p *Plugin) Interval(cfg PluginConfig) time.Duration {
	d := p.Recommended
	if cfg.Interval != "" {
		if v, err := time.ParseDuration(cfg.Interval); err == nil && v > 0 {
			d = v
		}
	}
	if d < p.MinInterval {
		d = p.MinInterval
	}
	if d > p.MaxInterval {
		d = p.MaxInterval
	}
	return d
}
