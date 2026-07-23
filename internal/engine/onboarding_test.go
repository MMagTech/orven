package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The first-run rules of CONSTRAINTS.md §28: the bundled demo is
// seed-enabled exactly once on a fresh installation and collects
// immediately; onboarding is permanent state that never reactivates;
// a restored demo returns disabled; a data directory from before the
// marker existed is never treated as a first run.

// seedEngine builds an engine whose seed directory carries the real
// bundled demo, against fresh data and plugin directories.
func seedEngine(t *testing.T, dataDir, pluginsDir string) *Engine {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	seedDir := filepath.Join(t.TempDir(), "seed")
	if err := copyDir(filepath.Join(root, "plugins", "demo-activity"),
		filepath.Join(seedDir, "demo-activity")); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	e := New(store, pluginsDir)
	e.SeedDir = seedDir
	e.SeedOnce()
	return e
}

func TestSeedEnableAndOnboardingLifecycle(t *testing.T) {
	dataDir, pluginsDir := t.TempDir(), t.TempDir()

	e := seedEngine(t, dataDir, pluginsDir)
	if e.Plugin("demo-activity") == nil {
		t.Fatal("fresh installation must be seeded with the demo")
	}
	if !e.Store.PluginConfig("demo-activity").Enabled {
		t.Fatal("fresh installation must seed-enable the demo (§28)")
	}
	if !e.OnboardingActive() {
		t.Fatal("fresh installation must start onboarding")
	}

	// The immediate first collection runs in the background; the first
	// Prepare must have material, so wait for it.
	deadline := time.Now().Add(15 * time.Second)
	for {
		if attempt, _ := e.Store.LastRun("demo-activity"); attempt != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("seed-enabled demo never collected")
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Restart: nothing re-seeds, nothing re-enables, onboarding holds.
	e2 := seedEngine(t, dataDir, pluginsDir)
	if !e2.Store.PluginConfig("demo-activity").Enabled || !e2.OnboardingActive() {
		t.Fatal("restart must not disturb seed-enable or onboarding state")
	}

	// The user disables the demo: onboarding ends, permanently.
	cfg := e2.Store.PluginConfig("demo-activity")
	cfg.Enabled = false
	e2.Store.SavePluginConfig("demo-activity", cfg)
	e2.FinishOnboarding()
	if e2.OnboardingActive() {
		t.Fatal("finishing onboarding must stick")
	}
	e3 := seedEngine(t, dataDir, pluginsDir)
	if e3.OnboardingActive() {
		t.Fatal("onboarding must never reactivate after a restart")
	}
	if e3.Store.PluginConfig("demo-activity").Enabled {
		t.Fatal("a disabled demo must never be re-enabled (§28)")
	}

	// The deliberate restore returns the demo installed but disabled.
	if err := e3.Uninstall("demo-activity", true, false); err != nil {
		t.Fatal(err)
	}
	if err := e3.RestoreSeed("demo-activity"); err != nil {
		t.Fatal(err)
	}
	if e3.Plugin("demo-activity") == nil {
		t.Fatal("restore must reinstall the demo")
	}
	if e3.Store.PluginConfig("demo-activity").Enabled {
		t.Fatal("a restored demo must return disabled (§28)")
	}
}

func TestOnboardingAbsentIsNotAFirstRun(t *testing.T) {
	e := testEngine(t) // data directory with no onboarding marker
	if e.OnboardingActive() {
		t.Fatal("no marker must mean no onboarding")
	}
	e.FinishOnboarding() // must be a no-op, not create state
	if _, err := os.Stat(e.onboardingPath()); err == nil {
		t.Fatal("finishing absent onboarding must not create the marker")
	}
}

func TestOnboardingRidesBackups(t *testing.T) {
	a := testEngine(t)
	a.startOnboarding()
	archive := filepath.Join(t.TempDir(), "a.zip")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.WriteBackup(f, false, ""); err != nil {
		t.Fatal(err)
	}
	f.Close()

	b := testEngine(t) // a machine with no onboarding state of its own
	if err := b.RestoreBackup(archive, ""); err != nil {
		t.Fatal(err)
	}
	if !b.OnboardingActive() {
		t.Fatal("restore must reproduce the backed-up onboarding state")
	}
}
