package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"orven/contract"
)

// TestThreeBriefingStates pins the product's three briefing states with
// two enabled sources:
//  1. complete coverage, nothing observed  -> the only all-clear
//  2. complete coverage, changes observed  -> no all-clear, changes shown
//  3. coverage incomplete                  -> neither claim; partial
//     briefing still carries everything successfully collected
func TestThreeBriefingStates(t *testing.T) {
	pluginsDir := t.TempDir()
	for _, id := range []string{"alpha", "beta"} {
		dir := filepath.Join(pluginsDir, id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		manifest := fmt.Sprintf(`schema_version: 1
id: %s
name: %s
version: 0.0.1
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
`, id, id)
		if err := os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := New(store, pluginsDir)
	for _, id := range []string{"alpha", "beta"} {
		e.Store.SavePluginConfig(id, PluginConfig{Enabled: true, Values: map[string]any{}})
	}

	save := func(id, status, summary string, items ...contract.Observation) {
		t.Helper()
		time.Sleep(20 * time.Millisecond)
		if err := e.Store.SaveBatch(StoredBatch{
			PluginID: id, PluginName: id, Collected: time.Now(),
			Status: status, Summary: summary, Items: items,
		}); err != nil {
			t.Fatal(err)
		}
	}
	generate := func() Brief {
		t.Helper()
		time.Sleep(20 * time.Millisecond)
		b, err := e.GenerateBrief()
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	// State 1: both checked, nothing anywhere.
	save("alpha", contract.StatusNothing, "No new activity.")
	save("beta", contract.StatusNothing, "No new activity.")
	b1 := generate()
	if !b1.CoverageComplete || !b1.Quiet {
		t.Fatalf("state 1 must be complete+quiet, got complete=%v quiet=%v", b1.CoverageComplete, b1.Quiet)
	}

	// State 2: both checked, alpha observed a change.
	save("alpha", contract.StatusOK, "1 new item.",
		contract.Observation{Title: "Backup completed", Scope: contract.ScopeEvent})
	save("beta", contract.StatusNothing, "No new activity.")
	b2 := generate()
	if !b2.CoverageComplete {
		t.Fatal("state 2: coverage must be complete when every source checked successfully")
	}
	if b2.Quiet {
		t.Fatal("state 2: an observed change must forfeit the all-clear")
	}

	// State 3: alpha observed a change, beta could not be checked. The
	// briefing keeps alpha's information and drops both normal claims.
	save("alpha", contract.StatusOK, "1 new item.",
		contract.Observation{Title: "Certificate renewed", Scope: contract.ScopeEvent})
	save("beta", contract.StatusUnavailable, "beta could not be reached.")
	b3 := generate()
	if b3.CoverageComplete || b3.Quiet {
		t.Fatalf("state 3 must be incomplete and not quiet, got complete=%v quiet=%v", b3.CoverageComplete, b3.Quiet)
	}
	found := false
	for _, s := range b3.Sections {
		if s.PluginID == "alpha" {
			for _, o := range s.Items {
				if o.Title == "Certificate renewed" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("state 3: a failed source must not discard information collected from the others")
	}
}
