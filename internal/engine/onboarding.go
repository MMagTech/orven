package engine

import (
	"os"
	"path/filepath"
	"strings"
)

// Onboarding is the first-run teaching state: it exists so a brand-new
// installation's Today page can explain what a Brief is and lead to
// the first one. The state is a marker file in the data directory —
// permanent, and carried through backups (see backupFiles) so a
// restored installation is never mistaken for a first run. A data
// directory from before the marker existed has no onboarding: absence
// means "not a first run", never "start onboarding".
const onboardingMarker = "onboarding"

func (e *Engine) onboardingPath() string {
	return filepath.Join(e.Store.Root, onboardingMarker)
}

// OnboardingActive reports whether the first-run experience should
// still be shown.
func (e *Engine) OnboardingActive() bool {
	b, err := os.ReadFile(e.onboardingPath())
	return err == nil && strings.TrimSpace(string(b)) == "active"
}

// FinishOnboarding permanently ends the first-run experience. It is
// called when the user demonstrates they no longer need it: a real
// plugin was enabled, or the demonstration plugin was disabled or
// uninstalled. It never reactivates.
func (e *Engine) FinishOnboarding() {
	if !e.OnboardingActive() {
		return
	}
	os.WriteFile(e.onboardingPath(), []byte("done\n"), 0o644)
}

func (e *Engine) startOnboarding() {
	os.WriteFile(e.onboardingPath(), []byte("active\n"), 0o644)
}
