package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Seed-once: the container image carries sample content (the demo
// plugin) outside the live plugins directory. On the first start of a
// fresh installation it is installed once, as a normal managed plugin,
// and a marker in the persistent data directory remembers that this
// happened — so uninstalling the demo sticks across restarts and
// container updates, and it can only come back through the deliberate
// RestoreSeed action.

const seedMarker = "seeded"

// SeedOnce installs every plugin from the seed directory into a fresh
// installation, exactly once per data directory.
func (e *Engine) SeedOnce() {
	if e.SeedDir == "" {
		return
	}
	marker := filepath.Join(e.Store.Root, seedMarker)
	if _, err := os.Stat(marker); err == nil {
		return // this installation was already seeded; never re-seed
	}
	entries, _ := os.ReadDir(e.SeedDir)
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if err := e.RestoreSeed(ent.Name()); err != nil {
			e.Logf("seeding %s failed: %v", ent.Name(), err)
		}
	}
	os.WriteFile(marker, []byte(time.Now().Format(time.RFC3339)+"\n"), 0o644)

	// Seed-enablement (CONSTRAINTS.md §28): the demonstration plugin —
	// bundled, first-party, zero permissions, fixture-only — is enabled
	// on a genuinely fresh installation and collects once immediately,
	// so the first Prepare has material to compile. This is the sole
	// exception to "nothing enables without the user's action"; it
	// happens only inside this fresh-install branch, so a demo the
	// user disables or uninstalls is never re-enabled.
	if p := e.Plugin(seedEnabledID); p != nil {
		cfg := e.Store.PluginConfig(seedEnabledID)
		cfg.Enabled = true
		e.Store.SavePluginConfig(seedEnabledID, cfg)
		e.startOnboarding()
		go e.TryRun(p, false)
	}
}

// seedEnabledID names the one plugin §28 permits to be seed-enabled.
// The manifest's permission declarations are prose, so the constraint
// is enforced by naming rather than inspection.
const seedEnabledID = "demo-activity"

// SeedAvailable reports whether id exists as seed content and is not
// currently installed — i.e. whether RestoreSeed would succeed.
func (e *Engine) SeedAvailable(id string) bool {
	if e.SeedDir == "" || e.Plugin(id) != nil {
		return false
	}
	_, err := os.Stat(filepath.Join(e.SeedDir, id, "plugin.yaml"))
	return err == nil
}

// RestoreSeed deliberately (re)installs one plugin from the seed
// directory as a managed install.
func (e *Engine) RestoreSeed(id string) error {
	if e.SeedDir == "" {
		return fmt.Errorf("this installation has no bundled seed content")
	}
	src := filepath.Join(e.SeedDir, id)
	if _, err := os.Stat(filepath.Join(src, "plugin.yaml")); err != nil {
		return fmt.Errorf("no bundled plugin named %q", id)
	}
	if e.Plugin(id) != nil {
		return fmt.Errorf("plugin %q is already installed", id)
	}
	dest := filepath.Join(e.PluginsDir, id)
	if err := os.MkdirAll(e.PluginsDir, 0o755); err != nil {
		return err
	}
	if err := copyDir(src, dest); err != nil {
		os.RemoveAll(dest)
		return err
	}
	p := LoadPlugin(dest)
	rec := InstallRecord{
		PluginID: id, Catalog: "bundled", Status: "bundled",
		Installed: time.Now(), Managed: true,
	}
	if p != nil {
		rec.Publisher = p.Manifest.Publisher
		rec.Version = p.Manifest.Version
	}
	if err := e.Store.SaveInstallRecord(rec); err != nil {
		return err
	}
	e.Reload()
	e.Logf("bundled plugin %s installed", id)
	return nil
}
