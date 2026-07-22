package engine

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"orven/contract"
)

// maxConcurrentRuns caps how many plugin subprocesses execute at once.
// Orven usually shares a small box with the systems it observes; when
// more plugins are due than this, the extras wait for a slot rather
// than spiking the host. Deliberately a constant — make it a setting
// only when someone demonstrates a real need.
const maxConcurrentRuns = 4

// Engine owns plugin execution, observations, and brief compilation.
// The core app (UI, settings pages) talks to it only through this type,
// so either side can change without breaking the other.
type Engine struct {
	Store      *Store
	PluginsDir string
	SeedDir    string // bundled sample content (ORVEN_SEED); "" outside the container

	mu      sync.Mutex
	running map[string]bool // prevents overlapping runs
	plugins []*Plugin
	sem     chan struct{} // bounds simultaneous plugin executions

	logMu  sync.Mutex
	logBuf []string
}

func New(store *Store, pluginsDir string) *Engine {
	e := &Engine{
		Store:      store,
		PluginsDir: pluginsDir,
		running:    map[string]bool{},
		sem:        make(chan struct{}, maxConcurrentRuns),
	}
	e.Reload()
	return e
}

// Reload re-discovers plugins from disk.
func (e *Engine) Reload() {
	ps := LoadPlugins(e.PluginsDir)
	e.mu.Lock()
	e.plugins = ps
	e.mu.Unlock()
	e.Logf("loaded %d plugin(s)", len(ps))
}

func (e *Engine) Plugins() []*Plugin {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]*Plugin(nil), e.plugins...)
}

func (e *Engine) Plugin(id string) *Plugin {
	for _, p := range e.Plugins() {
		if p.Manifest.ID == id {
			return p
		}
	}
	return nil
}

// TryRun runs a plugin unless it is already running (overlap guard).
func (e *Engine) TryRun(p *Plugin, manual bool) (StoredBatch, error) {
	if p.LoadError != "" || p.Incompatible {
		return StoredBatch{}, fmt.Errorf("plugin cannot run: %s", firstNonEmpty(p.LoadError, "incompatible with this engine"))
	}
	e.mu.Lock()
	if e.running[p.Manifest.ID] {
		e.mu.Unlock()
		return StoredBatch{}, fmt.Errorf("plugin is already running")
	}
	e.running[p.Manifest.ID] = true
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		delete(e.running, p.Manifest.ID)
		e.mu.Unlock()
	}()
	e.sem <- struct{}{} // wait for an execution slot
	defer func() { <-e.sem }()
	return e.Run(p, manual), nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ---- plugin health ----

// Health states, per the product spec: a plugin must never simply
// appear to do nothing.
const (
	HealthNeverConfigured = "Never configured"
	HealthReady           = "Ready"
	HealthWaiting         = "Waiting for next scheduled run"
	HealthRunning         = "Running"
	HealthHealthy         = "Healthy"
	HealthPartial         = "Partial data"
	HealthUnavailable     = "Source unavailable"
	HealthAuthFailed      = "Authentication failed"
	HealthTimedOut        = "Timed out"
	HealthFailed          = "Failed"
	HealthDisabled        = "Disabled"
	HealthIncompatible    = "Incompatible with current engine"
)

func (e *Engine) Health(p *Plugin) string {
	if p.LoadError != "" || p.Incompatible {
		return HealthIncompatible
	}
	cfg := e.Store.PluginConfig(p.Manifest.ID)
	if !cfg.Enabled {
		return HealthDisabled
	}
	e.mu.Lock()
	runningNow := e.running[p.Manifest.ID]
	e.mu.Unlock()
	if runningNow {
		return HealthRunning
	}
	attempt, _ := e.Store.LastRun(p.Manifest.ID)
	if attempt == nil {
		if e.requiredConfigMissing(p, cfg) {
			return HealthNeverConfigured
		}
		return HealthReady
	}
	switch attempt.Status {
	case contract.StatusOK, contract.StatusNothing:
		return HealthHealthy
	case contract.StatusPartial:
		return HealthPartial
	case contract.StatusUnavailable:
		return HealthUnavailable
	case contract.StatusAuthFailed:
		return HealthAuthFailed
	case "timeout":
		return HealthTimedOut
	default:
		return HealthFailed
	}
}

// requiredConfigMissing reports whether any required field without a
// default is still unset. Secret-type fields live in the secrets store,
// not in config values — checking the wrong place would leave a plugin
// "Never configured" forever after its API key was saved.
func (e *Engine) requiredConfigMissing(p *Plugin, cfg PluginConfig) bool {
	var secrets map[string]string
	for _, f := range p.Manifest.Config {
		if !f.Required || f.Default != nil {
			continue
		}
		if f.Type == "secret" {
			if secrets == nil {
				secrets = e.Store.Secrets(p.Manifest.ID)
			}
			if secrets[f.Key] == "" {
				return true
			}
			continue
		}
		if v, ok := cfg.Values[f.Key]; !ok || v == "" {
			return true
		}
	}
	return false
}

// ---- engine log (in-memory ring, mirrored to stderr) ----

const maxLogLines = 500

func (e *Engine) Logf(format string, args ...any) {
	line := time.Now().Format("2006-01-02 15:04:05") + "  " + fmt.Sprintf(format, args...)
	e.logMu.Lock()
	e.logBuf = append(e.logBuf, line)
	if len(e.logBuf) > maxLogLines {
		e.logBuf = e.logBuf[len(e.logBuf)-maxLogLines:]
	}
	e.logMu.Unlock()
	fmt.Fprintln(os.Stderr, "orven: "+line)
}

func (e *Engine) LogLines() []string {
	e.logMu.Lock()
	defer e.logMu.Unlock()
	return append([]string(nil), e.logBuf...)
}

// minimalEnv gives plugins only what a subprocess needs to start on the
// host OS — never the application's own environment wholesale.
func minimalEnv() []string {
	keep := []string{"PATH", "HOME", "LANG", "TZ",
		// Windows essentials; harmless elsewhere
		"SYSTEMROOT", "SYSTEMDRIVE", "USERPROFILE", "TEMP", "TMP", "COMSPEC", "PATHEXT", "LOCALAPPDATA"}
	var env []string
	for _, k := range keep {
		if v, ok := lookupEnvFold(k); ok {
			env = append(env, k+"="+v)
		}
	}
	return env
}

func lookupEnvFold(key string) (string, bool) {
	if v, ok := os.LookupEnv(key); ok {
		return v, true
	}
	for _, kv := range os.Environ() {
		if i := strings.IndexByte(kv, '='); i > 0 && strings.EqualFold(kv[:i], key) {
			return kv[i+1:], true
		}
	}
	return "", false
}
