package engine

import (
	"context"
	"time"
)

// StartScheduler runs the engine's single scheduling loop. Plugins never
// schedule themselves; this loop decides when each enabled plugin
// collects and when a briefing is assembled.
func (e *Engine) StartScheduler(ctx context.Context) {
	go func() {
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()
		lastBriefCheck := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-tick.C:
				e.collectDue(now)
				e.briefIfDue(lastBriefCheck, now)
				lastBriefCheck = now
			}
		}
	}()
}

func (e *Engine) collectDue(now time.Time) {
	for _, p := range e.Plugins() {
		if p.LoadError != "" || p.Incompatible {
			continue
		}
		cfg := e.Store.PluginConfig(p.Manifest.ID)
		if !cfg.Enabled || e.requiredConfigMissing(p, cfg) {
			continue
		}
		attempt, _ := e.Store.LastRun(p.Manifest.ID)
		if attempt != nil && now.Sub(attempt.Started) < p.Interval(cfg) {
			continue
		}
		go e.TryRun(p, false) // TryRun holds the overlap guard
	}
}

// briefIfDue fires when the configured local briefing time falls between
// the previous check and now, on an allowed day.
func (e *Engine) briefIfDue(prev, now time.Time) {
	s := e.Store.Settings()
	t, err := time.ParseInLocation("15:04", s.BriefTime, now.Location())
	if err != nil {
		return
	}
	due := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	if !due.After(prev) || due.After(now) {
		return
	}
	if len(s.BriefDays) > 0 && !contains(s.BriefDays, now.Format("Mon")) {
		return
	}
	if _, err := e.GenerateBrief(); err != nil {
		e.Logf("scheduled briefing failed: %v", err)
	}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
