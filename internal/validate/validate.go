// Package validate implements `orven validate`, the plugin validator
// specified in docs/VALIDATOR.md. It checks the manifest, runs the
// plugin against its own fixtures through the real engine runner, and
// inspects the output. It reports findings; it never rewrites plugin
// output, and any suggestion it displays differs from the original by
// capitalization or trailing punctuation only.
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"orven/contract"
	"orven/internal/engine"
)

type Finding struct {
	Severity   string // "ERROR" or "WARN"
	Where      string
	Message    string
	Suggestion string // display-only example; never applied
}

type report struct{ findings []Finding }

func (r *report) errf(where, format string, args ...any) {
	r.findings = append(r.findings, Finding{Severity: "ERROR", Where: where, Message: fmt.Sprintf(format, args...)})
}

func (r *report) warnf(where, format string, args ...any) {
	r.findings = append(r.findings, Finding{Severity: "WARN", Where: where, Message: fmt.Sprintf(format, args...)})
}

func (r *report) warnSuggest(where, msg, suggestion string) {
	r.findings = append(r.findings, Finding{Severity: "WARN", Where: where, Message: msg, Suggestion: suggestion})
}

// knownFieldTypes are the config field types the settings form renders.
var knownFieldTypes = map[string]bool{
	"text": true, "number": true, "boolean": true, "url": true,
	"duration": true, "select": true, "secret": true,
}

// knownKinds bounds the advisory `kind` vocabulary so it cannot drift
// into fifty private interpretations before it ever gains semantics.
var knownKinds = map[string]bool{
	"fact": true, "count": true, "change": true, "notice": true,
}

// forbiddenVoice is the maintained list of recommendation/remediation
// phrasing (VALIDATOR.md §9). It favors precision over recall: factual
// past-tense wording like "restarted 4 times" must not be flagged, so
// only clearly imperative or advisory forms appear here.
var forbiddenVoice = []string{
	"you should", "we recommend", "it is recommended", "recommended that",
	"consider ", "please ", "to fix", "try restarting", "try reinstalling",
	"needs to be restarted", "should be restarted", "you need to", "make sure to",
}

// Dir validates one plugin directory and returns all findings.
func Dir(dir string) []Finding {
	r := &report{}
	mfPath := filepath.Join(dir, "plugin.yaml")
	raw, err := os.ReadFile(mfPath)
	if err != nil {
		r.errf("plugin.yaml", "missing or unreadable: %v", err)
		return r.findings
	}
	_ = raw

	p := engine.LoadPlugin(dir)
	if p == nil {
		r.errf("plugin.yaml", "missing")
		return r.findings
	}
	if p.LoadError != "" {
		r.errf("plugin.yaml", "%s", p.LoadError)
		return r.findings
	}
	checkManifest(r, p)
	checkLayout(r, dir)
	checkFixtures(r, dir)

	if p.Incompatible {
		return r.findings // cannot run it here
	}
	runAndCheckOutput(r, p, dir)
	return r.findings
}

// ---- manifest checks (VALIDATOR.md §1-3, 7, 11-12) ----

func checkManifest(r *report, p *engine.Plugin) {
	m := p.Manifest
	if m.SchemaVersion != 1 {
		r.errf("plugin.yaml", "schema_version must be 1 (got %d)", m.SchemaVersion)
	}
	if m.Name == "" {
		r.errf("plugin.yaml", "name is required")
	}
	if m.Version == "" {
		r.errf("plugin.yaml", "version is required")
	}
	if m.Engine.MinContract > contract.Version {
		r.errf("plugin.yaml", "engine.min_contract %d is newer than this engine's contract %d", m.Engine.MinContract, contract.Version)
	}

	durs := map[string]string{
		"collection.recommended_interval": m.Collection.RecommendedInterval,
		"collection.min_interval":         m.Collection.MinInterval,
		"collection.max_interval":         m.Collection.MaxInterval,
		"collection.freshness":            m.Collection.Freshness,
		"timeout":                         m.Timeout,
	}
	parsed := map[string]time.Duration{}
	for field, v := range durs {
		if v == "" {
			continue
		}
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			r.errf("plugin.yaml", "%s: %q is not a valid duration (use forms like 30m, 2h)", field, v)
			continue
		}
		parsed[field] = d
	}
	if lo, hi := parsed["collection.min_interval"], parsed["collection.max_interval"]; lo > 0 && hi > 0 && lo > hi {
		r.errf("plugin.yaml", "collection.min_interval %s is greater than max_interval %s", lo, hi)
	}
	if m.Collection.Freshness == "" {
		r.warnf("plugin.yaml", "collection.freshness is not declared; the engine will assume 2x the recommended interval")
	}
	if len(m.Permissions) == 0 {
		r.warnf("plugin.yaml", "no permissions declared — every plugin touches something; say what")
	}

	seen := map[string]bool{}
	for _, f := range m.Config {
		where := fmt.Sprintf("plugin.yaml: config field %q", f.Key)
		if f.Key == "" {
			r.errf("plugin.yaml", "config field with empty key")
			continue
		}
		if seen[f.Key] {
			r.errf(where, "duplicate key")
		}
		seen[f.Key] = true
		if !knownFieldTypes[f.Type] {
			r.errf(where, "unknown type %q (known: text, number, boolean, url, duration, select, secret)", f.Type)
			continue
		}
		if f.Type == "select" && len(f.Options) == 0 {
			r.errf(where, "select field has no options")
		}
		checkDefault(r, where, f)
	}
}

func checkDefault(r *report, where string, f contract.ConfigField) {
	if f.Default == nil {
		return
	}
	switch f.Type {
	case "number":
		switch f.Default.(type) {
		case int, int64, float64:
		default:
			r.errf(where, "default %v does not match type number", f.Default)
		}
	case "boolean":
		if _, ok := f.Default.(bool); !ok {
			r.errf(where, "default %v does not match type boolean", f.Default)
		}
	case "select":
		s, ok := f.Default.(string)
		if !ok || !contains(f.Options, s) {
			r.errf(where, "default %v is not one of the declared options", f.Default)
		}
	case "secret":
		r.errf(where, "secret fields must not declare defaults")
	default:
		if _, ok := f.Default.(string); !ok {
			r.errf(where, "default %v does not match type %s", f.Default, f.Type)
		}
	}
}

// ---- layout checks (VALIDATOR.md §10) ----

func checkLayout(r *report, dir string) {
	if _, err := os.Stat(filepath.Join(dir, "README.md")); err != nil {
		r.warnf("layout", "README.md is missing — say what the plugin observes and what it needs")
	}
	for _, sub := range []string{"fixtures", "tests"} {
		entries, err := os.ReadDir(filepath.Join(dir, sub))
		if err != nil || len(entries) == 0 {
			r.warnf("layout", "%s/ is missing or empty — plugins must be testable without the real external system", sub)
		}
	}
}

// ---- execution and output checks (VALIDATOR.md §4-9, 13-17) ----

func runAndCheckOutput(r *report, p *engine.Plugin, dir string) {
	in := contract.Input{
		ContractVersion: contract.Version,
		PluginID:        p.Manifest.ID,
		Now:             time.Now().UTC(), // canonical UTC, matching the engine runner
		Config:          map[string]any{},
		Secrets:         map[string]string{},
	}
	var secretValues []string
	for _, f := range p.Manifest.Config {
		if f.Type == "secret" {
			v := "ORVEN-VALIDATE-SECRET-" + f.Key
			in.Secrets[f.Key] = v
			secretValues = append(secretValues, v)
			continue
		}
		if f.Default != nil {
			in.Config[f.Key] = f.Default
		} else if f.Required {
			in.Config[f.Key] = dummyFor(f)
		}
	}
	if fix := firstFixture(dir); fix != "" {
		in.Fixture = fix
	}

	out, err := engine.ExecPlugin(p, in)
	if err != nil {
		r.errf("execution", "%v", err)
		return
	}
	if out.ContractVersion == 0 {
		r.errf("output", "contract_version is missing")
	}
	if !legalStatus(out.Status) {
		r.errf("output", "status %q is not a contract status (ok, nothing, partial, unavailable, auth_failed, error)", out.Status)
	}
	if out.Status != contract.StatusOK && out.Summary == "" {
		r.warnf("output", "a %s result has no summary — failure and empty results should explain themselves", out.Status)
	}
	if strings.Count(strings.TrimSpace(out.Summary), ". ") >= 1 {
		r.warnf("output", "summary looks like more than one sentence — it should describe this one collection run briefly")
	}

	var textParts []string
	textParts = append(textParts, out.Summary, out.Error)
	for _, o := range out.Observations {
		where := fmt.Sprintf("output: observation %q", clip(o.Title, 40))
		if o.Title == "" {
			r.errf("output", "observation with empty title")
		}
		if o.Scope != "" && o.Scope != contract.ScopeEvent && o.Scope != contract.ScopeState {
			r.errf(where, "unknown scope %q (event, state, or omitted)", o.Scope)
		}
		if o.Kind != "" && !knownKinds[o.Kind] {
			r.warnf(where, "unknown kind %q — kind is advisory metadata with no engine behavior; known values are fact, count, change, notice", o.Kind)
		}
		titleStyle(r, o.Title)
		textParts = append(textParts, o.Title, o.Body)
	}
	all := strings.ToLower(strings.Join(textParts, "\n"))
	for _, phrase := range forbiddenVoice {
		if strings.Contains(all, phrase) {
			r.errf("output", "forbidden voice: output contains %q — state facts, never advise or instruct", strings.TrimSpace(phrase))
		}
	}
	blob, _ := json.Marshal(out)
	for _, sv := range secretValues {
		if strings.Contains(string(blob), sv) {
			r.errf("output", "secret leakage: a configured secret value appears in the plugin's output")
		}
	}
	// Credential-shaped content regardless of value (the runtime would
	// redact this; a plugin should never emit it in the first place).
	if frag := engine.ContainsCredentialPattern(strings.Join(textParts, "\n")); frag != "" {
		r.errf("output", "credential-shaped content (%q) — never put authorization headers or credential query parameters in output", clip(frag, 40))
	}
}

// checkFixtures warns when fixture files contain credential-shaped
// content — real credentials must never be committed to fixtures.
func checkFixtures(r *report, dir string) {
	entries, _ := os.ReadDir(filepath.Join(dir, "fixtures"))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, "fixtures", e.Name()))
		if err != nil || len(b) > 1<<20 {
			continue
		}
		if frag := engine.ContainsCredentialPattern(string(b)); frag != "" {
			r.warnf("fixtures/"+e.Name(), "credential-shaped content (%q) — never commit real credentials to fixtures; invent obviously fake values without header/parameter shapes", clip(frag, 40))
		}
	}
}

func dummyFor(f contract.ConfigField) any {
	switch f.Type {
	case "url":
		return "http://orven-validate.invalid"
	case "number":
		return 1
	case "boolean":
		return false
	case "duration":
		return "5m"
	case "select":
		return f.Options[0]
	default:
		return "orven-validate"
	}
}

func firstFixture(dir string) string {
	entries, err := os.ReadDir(filepath.Join(dir, "fixtures"))
	if err != nil {
		return ""
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return ""
	}
	sort.Strings(names)
	abs, err := filepath.Abs(filepath.Join(dir, "fixtures", names[0]))
	if err != nil {
		return ""
	}
	return abs
}

func legalStatus(s string) bool {
	switch s {
	case contract.StatusOK, contract.StatusNothing, contract.StatusPartial,
		contract.StatusUnavailable, contract.StatusAuthFailed, contract.StatusError:
		return true
	}
	return false
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
