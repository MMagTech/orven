package engine

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"orven/contract"
)

func testEngine(t *testing.T) *Engine {
	t.Helper()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return New(store, filepath.Join(root, "plugins"))
}

func TestLoadsDemoPlugin(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	if p == nil {
		t.Fatal("demo-activity not discovered")
	}
	if p.LoadError != "" {
		t.Fatalf("load error: %s", p.LoadError)
	}
	if p.Incompatible {
		t.Fatal("demo plugin marked incompatible with its own engine")
	}
	if p.Recommended.String() != "30m0s" {
		t.Fatalf("recommended interval = %s", p.Recommended)
	}
}

func TestIntervalClamping(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	if got := p.Interval(PluginConfig{Interval: "10s"}); got != p.MinInterval {
		t.Fatalf("below min not clamped: %s", got)
	}
	if got := p.Interval(PluginConfig{Interval: "9000h"}); got != p.MaxInterval {
		t.Fatalf("above max not clamped: %s", got)
	}
	if got := p.Interval(PluginConfig{}); got != p.Recommended {
		t.Fatalf("default should be recommended, got %s", got)
	}
}

func TestRunPluginAndGenerateBrief(t *testing.T) {
	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not on PATH")
	}
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})

	batch, err := e.TryRun(p, true)
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != contract.StatusOK {
		t.Fatalf("status = %s", batch.Status)
	}
	if len(batch.Items) == 0 {
		t.Fatal("no observations collected")
	}

	brief, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	if brief.Quiet {
		t.Fatal("brief should not be quiet after ok run")
	}
	if len(brief.Sections) != 1 || brief.Sections[0].PluginID != "demo-activity" {
		t.Fatalf("unexpected sections: %+v", brief.Sections)
	}
	if got, ok := e.Store.LatestBrief(); !ok || got.ID != brief.ID {
		t.Fatal("brief not retrievable from store")
	}
}

func TestEnabledPluginWithoutDataIsReportedAsMissing(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})

	brief, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	if len(brief.Sections) != 1 {
		t.Fatalf("expected a no-data section, got %+v", brief.Sections)
	}
	if brief.Sections[0].Status != "no_data" {
		t.Fatalf("missing coverage must be stated, got %q", brief.Sections[0].Status)
	}
	if brief.Quiet {
		t.Fatal("a plugin with no coverage means the environment is partly unknown — the brief must not claim quiet")
	}
}

// TestQuietRequiresEverySourceChecked: "All quiet" is only claimed when
// every enabled plugin checked successfully and found nothing. A source
// that could not be checked forfeits the all-clear.
func TestQuietRequiresEverySourceChecked(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})

	save := func(status, summary string) {
		t.Helper()
		time.Sleep(20 * time.Millisecond)
		if err := e.Store.SaveBatch(StoredBatch{
			PluginID: p.Manifest.ID, PluginName: p.Manifest.Name,
			Collected: time.Now(), Status: status, Summary: summary,
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

	// Case 1: checked successfully, nothing found -> All quiet.
	save(contract.StatusNothing, "No new activity.")
	if b := generate(); !b.Quiet {
		t.Fatal("every plugin checked and found nothing — brief must be quiet")
	}

	// Case 2: source unreachable -> no all-clear, and the section names
	// the source that could not be checked.
	save(contract.StatusUnavailable, "Demo media server could not be reached.")
	b := generate()
	if b.Quiet {
		t.Fatal("an unreachable source must forfeit the all-quiet claim")
	}
	if b.Sections[0].Summary != "Demo media server could not be reached." {
		t.Fatalf("the failed check must be clearly shown, got %q", b.Sections[0].Summary)
	}
}

// TestStateObservationLifecycle proves the agreed scope semantics with
// a failed import observed across three briefing windows:
//  1. observed repeatedly during one window  -> appears once in that briefing
//  2. still unresolved in the next window    -> appears once again
//  3. resolved (no longer observed)          -> absent from the next briefing
//
// Alongside it, an event from the first window must survive batch
// accumulation but never leak into later briefings.
func TestStateObservationLifecycle(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})

	failedImport := contract.Observation{
		Title: "1 import failed", Body: "The Marvels could not be imported.", Scope: contract.ScopeState,
	}
	backupDone := contract.Observation{
		Title: "Backup completed", Body: "Finished at 3:12 AM.", Scope: contract.ScopeEvent,
	}

	saveBatch := func(items ...contract.Observation) {
		t.Helper()
		time.Sleep(20 * time.Millisecond) // keep batch/brief timestamps strictly ordered
		err := e.Store.SaveBatch(StoredBatch{
			PluginID: p.Manifest.ID, PluginName: p.Manifest.Name,
			Collected: time.Now(), Status: contract.StatusOK, Items: items,
		})
		if err != nil {
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
	count := func(b Brief, title string) int {
		n := 0
		for _, s := range b.Sections {
			for _, o := range s.Items {
				if o.Title == title {
					n++
				}
			}
		}
		return n
	}

	// Window 1: three collections all see the failed import; one also
	// sees the backup event.
	saveBatch(backupDone, failedImport)
	saveBatch(failedImport)
	saveBatch(failedImport)
	b1 := generate()
	if got := count(b1, failedImport.Title); got != 1 {
		t.Fatalf("briefing 1: state observed 3 times must appear once, got %d", got)
	}
	if got := count(b1, backupDone.Title); got != 1 {
		t.Fatalf("briefing 1: event must survive accumulation, got %d", got)
	}

	// Window 2: still unresolved.
	saveBatch(failedImport)
	saveBatch(failedImport)
	b2 := generate()
	if got := count(b2, failedImport.Title); got != 1 {
		t.Fatalf("briefing 2: unresolved state must reappear exactly once, got %d", got)
	}
	if got := count(b2, backupDone.Title); got != 0 {
		t.Fatalf("briefing 2: old event must not leak into a later briefing, got %d", got)
	}

	// Window 3: resolved — the plugin checks successfully and no longer
	// observes the condition.
	saveBatch()
	b3 := generate()
	if got := count(b3, failedImport.Title); got != 0 {
		t.Fatalf("briefing 3: resolved state must disappear, got %d", got)
	}
}

// TestSummaryShownOnlyWithoutObservations: a batch summary describes
// one collection run, so a section that shows observations must not
// carry it; a section with nothing to show uses it as the explanation.
func TestSummaryShownOnlyWithoutObservations(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})

	save := func(status, summary string, items ...contract.Observation) {
		t.Helper()
		time.Sleep(20 * time.Millisecond)
		if err := e.Store.SaveBatch(StoredBatch{
			PluginID: p.Manifest.ID, PluginName: p.Manifest.Name,
			Collected: time.Now(), Status: status, Summary: summary, Items: items,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Window 1: observations present — summary must be suppressed.
	save(contract.StatusOK, "3 new items found.",
		contract.Observation{Title: "Backup completed", Scope: contract.ScopeEvent})
	time.Sleep(20 * time.Millisecond)
	b1, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	if got := b1.Sections[0].Summary; got != "" {
		t.Fatalf("section with observations must not carry a batch summary, got %q", got)
	}
	if len(b1.Sections[0].Items) != 1 {
		t.Fatalf("expected the observation to be the story, got %+v", b1.Sections[0].Items)
	}

	// Window 2: nothing to show — the summary is the explanation.
	save(contract.StatusUnavailable, "Demo media server could not be reached.")
	time.Sleep(20 * time.Millisecond)
	b2, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	if got := b2.Sections[0].Summary; got != "Demo media server could not be reached." {
		t.Fatalf("empty section must explain itself with the run summary, got %q", got)
	}
}

// TestStaleSectionIsMarked: a section whose data is older than the
// plugin's declared freshness window (2h for the demo plugin) must be
// marked stale; fresh data must not be.
func TestStaleSectionIsMarked(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})

	saveAt := func(collected time.Time) {
		t.Helper()
		if err := e.Store.SaveBatch(StoredBatch{
			PluginID: p.Manifest.ID, PluginName: p.Manifest.Name,
			Collected: collected, Status: contract.StatusOK,
			Items: []contract.Observation{{Title: "1 episode is stuck", Scope: contract.ScopeState}},
		}); err != nil {
			t.Fatal(err)
		}
	}

	saveAt(time.Now().Add(-3 * time.Hour)) // older than the 2h freshness window
	b1, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	if !b1.Sections[0].Stale {
		t.Fatal("data older than the declared freshness window must be marked stale")
	}

	time.Sleep(20 * time.Millisecond)
	saveAt(time.Now())
	time.Sleep(20 * time.Millisecond)
	b2, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	if b2.Sections[0].Stale {
		t.Fatal("freshly collected data must not be marked stale")
	}
}

func TestHealthStates(t *testing.T) {
	e := testEngine(t)
	p := e.Plugin("demo-activity")
	if h := e.Health(p); h != HealthDisabled {
		t.Fatalf("fresh plugin should be Disabled, got %s", h)
	}
	e.Store.SavePluginConfig(p.Manifest.ID, PluginConfig{Enabled: true, Values: map[string]any{}})
	if h := e.Health(p); h != HealthReady {
		t.Fatalf("enabled unrun plugin should be Ready, got %s", h)
	}
}
