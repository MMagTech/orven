// Package contract defines the versioned boundary between the briefing
// engine and plugins. Backwards compatibility rule: a plugin written
// against contract version N must keep working on every engine whose
// contract version is >= N. Fields may be added; they must never be
// removed or repurposed.
package contract

import "time"

// Version is the contract version this engine speaks.
const Version = 1

// Plugin result statuses. A plugin must always report one of these so
// that missing coverage is never mistaken for "everything is fine".
const (
	StatusOK          = "ok"           // relevant activity was found
	StatusNothing     = "nothing"      // checked successfully, nothing relevant
	StatusPartial     = "partial"      // only partial information was available
	StatusUnavailable = "unavailable"  // the source system was unavailable
	StatusAuthFailed  = "auth_failed"  // authentication failed
	StatusError       = "error"        // the plugin could not complete its check
)

// Input is what the engine writes to a plugin's stdin as JSON.
type Input struct {
	ContractVersion int               `json:"contract_version"`
	PluginID        string            `json:"plugin_id"`
	Now             time.Time         `json:"now"`
	WindowStart     time.Time         `json:"window_start"` // last successful run, or zero
	Config          map[string]any    `json:"config"`
	Secrets         map[string]string `json:"secrets,omitempty"`
	Fixture         string            `json:"fixture,omitempty"` // set during tests: path to fixture data
}

// Output is what a plugin writes to stdout as JSON.
type Output struct {
	ContractVersion int           `json:"contract_version"`
	Status          string        `json:"status"`
	// Summary is one factual sentence describing this collection run.
	// It appears in run history, and in a briefing only when there are
	// no observations to show (to explain why). It is never the
	// headline above observations — the engine composes the briefing.
	Summary string `json:"summary,omitempty"`
	Observations    []Observation `json:"observations,omitempty"`
	Error           string        `json:"error,omitempty"` // internal detail, shown in logs, never in briefs
}

// Observation scopes. The deciding question for plugin authors: if the
// condition resolves before the next briefing, should the reader still
// be told it happened? Yes -> event. No -> state. (Examples across
// domains: docs/PLUGIN_SDK.md.)
const (
	// ScopeEvent: something that occurred once. Events accumulate
	// across every collection in a briefing window so nothing that
	// happened between briefings is lost. This is the default.
	ScopeEvent = "event"
	// ScopeState: a condition that is currently true. Only the most
	// recent collection counts for a briefing, so a condition appears
	// once per briefing — and keeps appearing in later briefings for
	// as long as the plugin still observes it.
	ScopeState = "state"
)

// Observation is a single structured fact a plugin found. Plugins state
// facts only — never suggestions, fixes, or remediation steps.
type Observation struct {
	Title      string    `json:"title"`
	Body       string    `json:"body,omitempty"`
	Kind       string    `json:"kind,omitempty"`  // fact | count | change | notice
	Scope      string    `json:"scope,omitempty"` // event (default) | state
	OccurredAt time.Time `json:"occurred_at,omitempty"`
}

// Manifest is the parsed plugin.yaml.
type Manifest struct {
	SchemaVersion int      `yaml:"schema_version"`
	ID            string   `yaml:"id"`
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	Publisher     string   `yaml:"publisher"`
	Description   string   `yaml:"description"`
	Entrypoint    []string `yaml:"entrypoint"`
	Engine        struct {
		MinContract int `yaml:"min_contract"`
	} `yaml:"engine"`
	Collection struct {
		RecommendedInterval string `yaml:"recommended_interval"`
		MinInterval         string `yaml:"min_interval"`
		MaxInterval         string `yaml:"max_interval"`
		Freshness           string `yaml:"freshness"`
	} `yaml:"collection"`
	Timeout     string        `yaml:"timeout"`
	Permissions []string      `yaml:"permissions"`
	Config      []ConfigField `yaml:"config"`
}

// ConfigField describes one setting in a plugin's schema-generated form.
type ConfigField struct {
	Key      string   `yaml:"key"`
	Type     string   `yaml:"type"` // text | number | boolean | url | duration | select | secret
	Label    string   `yaml:"label"`
	Help     string   `yaml:"help"`
	Default  any      `yaml:"default"`
	Required bool     `yaml:"required"`
	Options  []string `yaml:"options"`
}
