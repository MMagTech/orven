package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"orven/contract"
)

// Store is file-based storage owned by the engine. Everything lives
// under one data directory so backups are a plain copy of that folder.
//
//	data/
//	  settings.json
//	  config/<plugin>.json     plugin configuration
//	  secrets/<plugin>.json    plugin credentials (write-only via UI)
//	  runs/<plugin>.json       execution history (bounded)
//	  observations/<plugin>/<stamp>.json
//	  briefs/<stamp>.json
type Store struct {
	Root string
}

func NewStore(root string) (*Store, error) {
	for _, d := range []string{"", "config", "secrets", "runs", "observations", "briefs"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			return nil, err
		}
	}
	return &Store{Root: root}, nil
}

func (s *Store) readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (s *Store) writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ---- app settings ----

type Settings struct {
	BriefTime     string   `json:"brief_time"`     // "07:30"
	BriefDays     []string `json:"brief_days"`     // ["Mon",...]; empty = every day
	RetentionDays int      `json:"retention_days"` // briefs + observations
	Repos         []string `json:"repos"`          // plugin repositories
}

func DefaultSettings() Settings {
	return Settings{
		BriefTime:     "07:30",
		BriefDays:     nil,
		RetentionDays: 90,
		Repos:         []string{"https://github.com/mmagnant/orven-plugins"},
	}
}

func (s *Store) Settings() Settings {
	var v Settings
	if err := s.readJSON(filepath.Join(s.Root, "settings.json"), &v); err != nil {
		return DefaultSettings()
	}
	if v.RetentionDays <= 0 {
		v.RetentionDays = 90
	}
	return v
}

func (s *Store) SaveSettings(v Settings) error {
	return s.writeJSON(filepath.Join(s.Root, "settings.json"), v)
}

// ---- plugin config ----

type PluginConfig struct {
	Enabled  bool           `json:"enabled"`
	Interval string         `json:"interval,omitempty"` // user override, e.g. "30m"
	Values   map[string]any `json:"values"`
}

func (s *Store) PluginConfig(id string) PluginConfig {
	var v PluginConfig
	if err := s.readJSON(filepath.Join(s.Root, "config", id+".json"), &v); err != nil {
		return PluginConfig{Values: map[string]any{}}
	}
	if v.Values == nil {
		v.Values = map[string]any{}
	}
	return v
}

func (s *Store) SavePluginConfig(id string, v PluginConfig) error {
	return s.writeJSON(filepath.Join(s.Root, "config", id+".json"), v)
}

// ---- plugin secrets (write-only: UI can set, test presence, remove) ----

func (s *Store) Secrets(id string) map[string]string {
	v := map[string]string{}
	s.readJSON(filepath.Join(s.Root, "secrets", id+".json"), &v)
	return v
}

func (s *Store) SaveSecrets(id string, v map[string]string) error {
	return s.writeJSON(filepath.Join(s.Root, "secrets", id+".json"), v)
}

// ---- run history ----

type RunRecord struct {
	Started  time.Time `json:"started"`
	Finished time.Time `json:"finished"`
	Status   string    `json:"status"` // contract status, or "timeout" / "invalid_output"
	Summary  string    `json:"summary,omitempty"`
	Error    string    `json:"error,omitempty"`
	Manual   bool      `json:"manual,omitempty"`
}

const maxRunHistory = 50

func (s *Store) Runs(id string) []RunRecord {
	var v []RunRecord
	s.readJSON(filepath.Join(s.Root, "runs", id+".json"), &v)
	return v
}

func (s *Store) AppendRun(id string, r RunRecord) error {
	runs := append(s.Runs(id), r)
	if len(runs) > maxRunHistory {
		runs = runs[len(runs)-maxRunHistory:]
	}
	return s.writeJSON(filepath.Join(s.Root, "runs", id+".json"), runs)
}

// LastRun returns the most recent attempt and most recent success.
func (s *Store) LastRun(id string) (attempt, success *RunRecord) {
	runs := s.Runs(id)
	for i := len(runs) - 1; i >= 0; i-- {
		r := runs[i]
		if attempt == nil {
			attempt = &r
		}
		if success == nil && (r.Status == contract.StatusOK || r.Status == contract.StatusNothing || r.Status == contract.StatusPartial) {
			success = &r
			break
		}
	}
	return
}

// ---- observations ----

type StoredBatch struct {
	PluginID   string                 `json:"plugin_id"`
	PluginName string                 `json:"plugin_name"`
	Collected  time.Time              `json:"collected"`
	Status     string                 `json:"status"`
	Summary    string                 `json:"summary,omitempty"`
	Items      []contract.Observation `json:"items,omitempty"`
}

const stamp = "20060102T150405.000"

func (s *Store) SaveBatch(b StoredBatch) error {
	dir := filepath.Join(s.Root, "observations", b.PluginID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return s.writeJSON(filepath.Join(dir, b.Collected.UTC().Format(stamp)+".json"), b)
}

// BatchesSince returns all stored batches collected after t, oldest first.
func (s *Store) BatchesSince(t time.Time) []StoredBatch {
	var out []StoredBatch
	root := filepath.Join(s.Root, "observations")
	dirs, _ := os.ReadDir(root)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		files, _ := os.ReadDir(filepath.Join(root, d.Name()))
		for _, f := range files {
			ts, err := time.Parse(stamp, strings.TrimSuffix(f.Name(), ".json"))
			if err != nil || !ts.After(t) {
				continue
			}
			var b StoredBatch
			if s.readJSON(filepath.Join(root, d.Name(), f.Name()), &b) == nil {
				out = append(out, b)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Collected.Before(out[j].Collected) })
	return out
}

// ---- briefs ----

// Brief distinguishes three states, and the first two claims are only
// ever made honestly:
//   - CoverageComplete && Quiet: every source was checked, nothing
//     changed — the only time the app states an all-clear.
//   - CoverageComplete && !Quiet: changes are presented without
//     classifying their importance; that judgment belongs to the reader.
//   - !CoverageComplete: some enabled source could not be (fully)
//     collected, so neither claim above can be made; the briefing leads
//     with that and then presents what was collected.
type Brief struct {
	ID               string         `json:"id"`
	Generated        time.Time      `json:"generated"`
	Window           time.Time      `json:"window_start"`
	Sections         []BriefSection `json:"sections"`
	Quiet            bool           `json:"quiet"`
	CoverageComplete bool           `json:"coverage_complete"`
}

type BriefSection struct {
	PluginID   string                 `json:"plugin_id"`
	PluginName string                 `json:"plugin_name"`
	Status     string                 `json:"status"`
	Summary    string                 `json:"summary,omitempty"`
	Items      []contract.Observation `json:"items,omitempty"`
	Freshness  time.Time              `json:"freshness"` // when this data was collected
	// Stale records that, at generation time, this section's data was
	// older than the plugin's declared freshness window.
	Stale bool `json:"stale,omitempty"`
}

func (s *Store) SaveBrief(b Brief) error {
	return s.writeJSON(filepath.Join(s.Root, "briefs", b.ID+".json"), b)
}

func (s *Store) Brief(id string) (Brief, error) {
	var b Brief
	if strings.ContainsAny(id, `/\.`) {
		return b, fmt.Errorf("bad brief id")
	}
	err := s.readJSON(filepath.Join(s.Root, "briefs", id+".json"), &b)
	return b, err
}

// Briefs returns all brief IDs, newest first.
func (s *Store) Briefs() []Brief {
	files, _ := os.ReadDir(filepath.Join(s.Root, "briefs"))
	var out []Brief
	for _, f := range files {
		var b Brief
		if s.readJSON(filepath.Join(s.Root, "briefs", f.Name()), &b) == nil {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Generated.After(out[j].Generated) })
	return out
}

func (s *Store) LatestBrief() (Brief, bool) {
	all := s.Briefs()
	if len(all) == 0 {
		return Brief{}, false
	}
	return all[0], true
}

// Prune removes briefs and observations older than the retention window.
func (s *Store) Prune(retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, b := range s.Briefs() {
		if b.Generated.Before(cutoff) {
			os.Remove(filepath.Join(s.Root, "briefs", b.ID+".json"))
		}
	}
	root := filepath.Join(s.Root, "observations")
	dirs, _ := os.ReadDir(root)
	for _, d := range dirs {
		files, _ := os.ReadDir(filepath.Join(root, d.Name()))
		for _, f := range files {
			ts, err := time.Parse(stamp, strings.TrimSuffix(f.Name(), ".json"))
			if err == nil && ts.Before(cutoff) {
				os.Remove(filepath.Join(root, d.Name(), f.Name()))
			}
		}
	}
}
