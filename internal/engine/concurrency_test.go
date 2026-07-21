package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"orven/contract"
)

// TestConcurrencyCeiling launches 6 slow plugins at once and proves,
// from the start/end times each subprocess records on disk, that no
// more than maxConcurrentRuns executed simultaneously and that every
// waiting plugin still completed normally.
func TestConcurrencyCeiling(t *testing.T) {
	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not on PATH")
	}

	const script = `import json, sys, time
json.load(sys.stdin)
with open("t_start", "w") as f:
    f.write(repr(time.time()))
time.sleep(0.7)
with open("t_end", "w") as f:
    f.write(repr(time.time()))
json.dump({"contract_version": 1, "status": "nothing", "summary": "checked"}, sys.stdout)
`
	const n = 6
	pluginsDir := t.TempDir()
	for i := 0; i < n; i++ {
		dir := filepath.Join(pluginsDir, fmt.Sprintf("slow-%d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		manifest := fmt.Sprintf(`schema_version: 1
id: slow-%d
name: Slow %d
version: 0.0.1
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
timeout: 30s
`, i, i)
		if err := os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte(script), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := New(store, pluginsDir)
	if got := len(e.Plugins()); got != n {
		t.Fatalf("expected %d plugins, loaded %d", n, got)
	}

	var wg sync.WaitGroup
	for _, p := range e.Plugins() {
		wg.Add(1)
		go func(p *Plugin) {
			defer wg.Done()
			batch, err := e.TryRun(p, true)
			if err != nil {
				t.Errorf("%s: %v", p.Manifest.ID, err)
				return
			}
			if batch.Status != contract.StatusNothing {
				t.Errorf("%s: expected clean completion, got %q", p.Manifest.ID, batch.Status)
			}
		}(p)
	}
	wg.Wait()

	// Reconstruct true concurrency from what the subprocesses recorded.
	type edge struct {
		at    float64
		delta int
	}
	var edges []edge
	for i := 0; i < n; i++ {
		dir := filepath.Join(pluginsDir, fmt.Sprintf("slow-%d", i))
		start := readFloat(t, filepath.Join(dir, "t_start"))
		end := readFloat(t, filepath.Join(dir, "t_end"))
		edges = append(edges, edge{start, +1}, edge{end, -1})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].at < edges[j].at })
	peak, cur := 0, 0
	for _, ev := range edges {
		cur += ev.delta
		if cur > peak {
			peak = cur
		}
	}
	if peak > maxConcurrentRuns {
		t.Fatalf("peak concurrency %d exceeded the ceiling of %d", peak, maxConcurrentRuns)
	}
	if peak < 2 {
		t.Fatalf("peak concurrency %d — the ceiling should stagger runs, not serialize them", peak)
	}
}

func readFloat(t *testing.T, path string) float64 {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("plugin did not record %s: %v", filepath.Base(path), err)
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
