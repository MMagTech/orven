package core

import (
	"strings"
	"testing"
	"time"

	"orven/contract"
	"orven/internal/engine"
)

func exportBrief() engine.Brief {
	return engine.Brief{
		ID:        "20260722T073000",
		Generated: time.Date(2026, 7, 22, 7, 30, 0, 0, time.Local),
		Sections: []engine.BriefSection{
			{
				PluginID: "sonarr", PluginName: "Sonarr", Status: contract.StatusOK,
				Items: []contract.Observation{
					{Title: "2 episodes downloaded overnight", Body: "Both ready to watch.", Scope: contract.ScopeEvent},
				},
				Freshness: time.Date(2026, 7, 22, 7, 0, 0, 0, time.Local),
			},
			{PluginID: "traefik", PluginName: "Traefik", Status: contract.StatusNothing},
			{
				PluginID: "backup", PluginName: "CrashPlan", Status: contract.StatusUnavailable,
				Summary: "The backup host could not be reached.",
			},
		},
	}
}

func TestBriefMarkdown(t *testing.T) {
	md := BriefMarkdown(exportBrief())
	for _, want := range []string{
		"# The Morning Brief — Wednesday, July 22, 2026",
		"> **Unable to verify all sources** — CrashPlan could not be checked.",
		"## Sonarr",
		"- **2 episodes downloaded overnight** — Both ready to watch.",
		"**Coverage**",
		"- This briefing draws on: Sonarr and Traefik.",
		"- Could not be checked: CrashPlan — The backup host could not be reached.",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q\n---\n%s", want, md)
		}
	}
	// facts-only rule survives export: no advisory language crept in
	for _, banned := range []string{"you should", "we recommend"} {
		if strings.Contains(strings.ToLower(md), banned) {
			t.Errorf("advisory language in export: %q", banned)
		}
	}
	// quiet section (Traefik) contributes to coverage but has no story
	if strings.Contains(md, "## Traefik") {
		t.Error("a nothing-status section must not become a story heading")
	}
}
