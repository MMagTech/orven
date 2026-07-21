package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"orven/contract"
)

const maxOutputBytes = 1 << 20 // 1 MiB of plugin stdout is plenty

// Run executes one plugin as a subprocess: config in on stdin, one JSON
// Output on stdout, hard timeout, no shell. The plugin gets a stripped
// environment and its own directory as working directory.
func (e *Engine) Run(p *Plugin, manual bool) StoredBatch {
	started := time.Now()
	cfg := e.Store.PluginConfig(p.Manifest.ID)

	_, lastOK := e.Store.LastRun(p.Manifest.ID)
	var window time.Time
	if lastOK != nil {
		window = lastOK.Started
	}

	in := contract.Input{
		ContractVersion: contract.Version,
		PluginID:        p.Manifest.ID,
		Now:             started,
		WindowStart:     window,
		Config:          e.configWithDefaults(p, cfg),
		Secrets:         e.Store.Secrets(p.Manifest.ID),
	}

	out, runErr := execPlugin(p, in)

	batch := StoredBatch{
		PluginID:   p.Manifest.ID,
		PluginName: p.Manifest.Name,
		Collected:  started,
	}
	rec := RunRecord{Started: started, Finished: time.Now(), Manual: manual}

	switch {
	case runErr != nil:
		rec.Status = statusForErr(runErr)
		rec.Error = runErr.Error()
		batch.Status = rec.Status
	case !validStatus(out.Status):
		rec.Status = "invalid_output"
		rec.Error = fmt.Sprintf("plugin reported unknown status %q", out.Status)
		batch.Status = contract.StatusError
	default:
		rec.Status = out.Status
		rec.Summary = out.Summary
		rec.Error = out.Error
		batch.Status = out.Status
		batch.Summary = out.Summary
		batch.Items = out.Observations
	}

	if err := e.Store.SaveBatch(batch); err != nil {
		log.Printf("engine: save batch for %s: %v", p.Manifest.ID, err)
	}
	if err := e.Store.AppendRun(p.Manifest.ID, rec); err != nil {
		log.Printf("engine: record run for %s: %v", p.Manifest.ID, err)
	}
	e.Logf("plugin %s finished: %s%s", p.Manifest.ID, rec.Status, errSuffix(rec.Error))
	return batch
}

func errSuffix(s string) string {
	if s == "" {
		return ""
	}
	return " (" + s + ")"
}

func statusForErr(err error) string {
	if strings.Contains(err.Error(), "timed out") {
		return "timeout"
	}
	return contract.StatusError
}

func validStatus(s string) bool {
	switch s {
	case contract.StatusOK, contract.StatusNothing, contract.StatusPartial,
		contract.StatusUnavailable, contract.StatusAuthFailed, contract.StatusError:
		return true
	}
	return false
}

// pythonAliases is the entire interpreter-fallback table: each standard
// Python launcher name and its one permitted substitute. Exact bare
// command names only, resolved with the normal system executable
// lookup — never a shell. Plugins cannot extend this table, and no
// other entrypoint command is ever altered.
var pythonAliases = map[string]string{"python": "python3", "python3": "python"}

// resolveInterpreter returns the command to actually execute for a
// declared entrypoint command. Only the two standard Python names are
// eligible: if the declared one is not on PATH and its sibling is, the
// sibling is used, so `python` manifests work on systems that only
// ship `python3` (and vice versa).
func resolveInterpreter(name string, lookPath func(string) (string, error)) string {
	alias, ok := pythonAliases[name]
	if !ok {
		return name
	}
	if _, err := lookPath(name); err == nil {
		return name
	}
	if _, err := lookPath(alias); err == nil {
		return alias
	}
	return name
}

func execPlugin(p *Plugin, in contract.Input) (contract.Output, error) {
	var out contract.Output

	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout)
	defer cancel()

	stdin, err := json.Marshal(in)
	if err != nil {
		return out, err
	}

	command := resolveInterpreter(p.Manifest.Entrypoint[0], exec.LookPath)
	cmd := exec.CommandContext(ctx, command, p.Manifest.Entrypoint[1:]...)
	cmd.Dir = p.Dir
	cmd.Env = minimalEnv()
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{w: &stdout, n: maxOutputBytes}
	cmd.Stderr = &limitedWriter{w: &stderr, n: 16 << 10}

	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("timed out after %s", p.Timeout)
	}
	if err != nil {
		return out, fmt.Errorf("plugin exited with error: %v: %s", err, firstLine(stderr.String()))
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return out, fmt.Errorf("plugin produced malformed output: %v", err)
	}
	return out, nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// configWithDefaults merges declared defaults under the saved values so
// a plugin always receives every declared key.
func (e *Engine) configWithDefaults(p *Plugin, cfg PluginConfig) map[string]any {
	m := map[string]any{}
	for _, f := range p.Manifest.Config {
		if f.Default != nil {
			m[f.Key] = f.Default
		}
	}
	for k, v := range cfg.Values {
		m[k] = v
	}
	return m
}

type limitedWriter struct {
	w interface{ Write([]byte) (int, error) }
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n <= 0 {
		return len(p), nil // swallow the rest
	}
	if len(p) > l.n {
		p = p[:l.n]
	}
	n, err := l.w.Write(p)
	l.n -= n
	return len(p), err
}
