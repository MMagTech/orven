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
// plugin.yaml. A broken manifest yields a Plugin with LoadError set so
// the UI can explain why it cannot run, rather than silently hiding it.
func LoadPlugins(dir string) []*Plugin {
	var out []*Plugin
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pdir := filepath.Join(dir, e.Name())
		mf := filepath.Join(pdir, "plugin.yaml")
		b, err := os.ReadFile(mf)
		if err != nil {
			continue // not a plugin folder
		}
		p := &Plugin{Dir: pdir}
		if err := yaml.Unmarshal(b, &p.Manifest); err != nil {
			p.Manifest.ID = e.Name()
			p.Manifest.Name = e.Name()
			p.LoadError = fmt.Sprintf("invalid plugin.yaml: %v", err)
			out = append(out, p)
			continue
		}
		if p.Manifest.ID == "" || len(p.Manifest.Entrypoint) == 0 {
			p.Manifest.ID = e.Name()
			if p.Manifest.Name == "" {
				p.Manifest.Name = e.Name()
			}
			p.LoadError = "plugin.yaml must declare id and entrypoint"
			out = append(out, p)
			continue
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
		out = append(out, p)
	}
	return out
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
