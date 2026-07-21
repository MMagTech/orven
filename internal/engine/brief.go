package engine

import (
	"time"

	"orven/contract"
)

// GenerateBrief compiles a briefing from stored observations. For each
// enabled plugin the most recent batch in the window provides the
// section's status. Observation items follow their scope: events
// accumulate across every batch in the window so nothing that happened
// is lost; states are taken from the newest batch only, so a condition
// appears once per briefing yet returns in every later briefing for as
// long as the plugin still observes it. Plugins with no coverage in
// the window are reported as such — absence of data is never presented
// as good news.
func (e *Engine) GenerateBrief() (Brief, error) {
	now := time.Now()
	window := now.Add(-14 * 24 * time.Hour) // hard floor
	if prev, ok := e.Store.LatestBrief(); ok && prev.Generated.After(window) {
		window = prev.Generated
	}

	batches := e.Store.BatchesSince(window)
	latest := map[string]StoredBatch{}           // plugin id -> newest batch
	items := map[string][]contract.Observation{} // events across window, then current states
	for _, b := range batches {
		latest[b.PluginID] = b
		for _, o := range b.Items {
			if o.Scope != contract.ScopeState {
				items[b.PluginID] = append(items[b.PluginID], o)
			}
		}
	}
	for id, b := range latest {
		for _, o := range b.Items {
			if o.Scope == contract.ScopeState {
				items[id] = append(items[id], o)
			}
		}
	}

	brief := Brief{
		ID:        now.UTC().Format("20060102T150405"),
		Generated: now,
		Window:    window,
	}

	for _, p := range e.Plugins() {
		cfg := e.Store.PluginConfig(p.Manifest.ID)
		if !cfg.Enabled || p.LoadError != "" || p.Incompatible {
			continue
		}
		sec := BriefSection{PluginID: p.Manifest.ID, PluginName: p.Manifest.Name}
		if b, ok := latest[p.Manifest.ID]; ok {
			sec.Status = b.Status
			sec.Items = items[p.Manifest.ID]
			sec.Freshness = b.Collected
			sec.Stale = now.Sub(b.Collected) > p.Freshness
			// A summary describes one collection run, so it can't lead a
			// section that aggregates a whole window. When there are
			// observations, they are the story; the summary appears only
			// when it must explain why there is nothing to show.
			if len(sec.Items) == 0 {
				sec.Summary = b.Summary
			}
		} else {
			// enabled but no coverage in this window
			sec.Status = "no_data"
			if attempt, _ := e.Store.LastRun(p.Manifest.ID); attempt == nil {
				sec.Summary = "This plugin has not run yet."
			} else {
				sec.Summary = "No information was collected for this period."
			}
		}
		brief.Sections = append(brief.Sections, sec)
	}

	// Coverage is complete only when every enabled source was checked
	// successfully and in full. "All quiet" is a positive claim on top
	// of that: complete coverage and nothing observed anywhere. A
	// failed, partial, or missing check means part of the environment
	// is unknown — unknown is never quiet, and incomplete coverage
	// forfeits both normal opening claims.
	brief.CoverageComplete = len(brief.Sections) > 0
	totalItems := 0
	for _, s := range brief.Sections {
		if s.Status != contract.StatusOK && s.Status != contract.StatusNothing {
			brief.CoverageComplete = false
		}
		totalItems += len(s.Items)
	}
	brief.Quiet = brief.CoverageComplete && totalItems == 0

	if err := e.Store.SaveBrief(brief); err != nil {
		return brief, err
	}
	e.Logf("briefing %s generated: %d section(s), quiet=%v", brief.ID, len(brief.Sections), brief.Quiet)
	e.Store.Prune(e.Store.Settings().RetentionDays)
	return brief, nil
}
