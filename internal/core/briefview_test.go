package core

import (
	"fmt"
	"testing"
	"time"

	"orven/contract"
	"orven/internal/engine"
)

// The projection rules of the reading contract (CONSTRAINTS.md §27):
// checked-and-quiet sections collapse into "Also checked"; anything
// unchecked, stale, or partial is itself news and stays a story; the
// item cap folds only what the page leads with, never the record.

func TestBriefViewPartition(t *testing.T) {
	b := engine.Brief{
		Generated: time.Now(),
		Sections: []engine.BriefSection{
			{PluginName: "Sonarr", Status: contract.StatusOK,
				Items: []contract.Observation{{Title: "1 episode downloaded"}}},
			{PluginName: "Traefik", Status: contract.StatusNothing},
			{PluginName: "Backups", Status: contract.StatusUnavailable,
				Summary: "The backup host could not be reached."},
			{PluginName: "Certs", Status: contract.StatusNothing, Stale: true,
				Summary: "Nothing new.", Freshness: time.Now().Add(-3 * time.Hour)},
		},
	}
	view := briefView(b)
	stories := view["Stories"].([]storyView)

	var names []string
	for _, s := range stories {
		names = append(names, s.PluginName)
	}
	want := []string{"Sonarr", "Backups", "Certs"}
	if len(names) != len(want) {
		t.Fatalf("stories = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("stories = %v, want %v", names, want)
		}
	}
	// Only the fresh, successfully checked, quiet source collapses.
	if got := view["AlsoChecked"].(string); got != "Traefik" {
		t.Errorf("AlsoChecked = %q, want %q", got, "Traefik")
	}
}

func TestBriefViewQuietSuppressesAlsoChecked(t *testing.T) {
	b := engine.Brief{
		Generated: time.Now(),
		Quiet:     true, CoverageComplete: true,
		Sections: []engine.BriefSection{
			{PluginName: "Traefik", Status: contract.StatusNothing},
			{PluginName: "Sonarr", Status: contract.StatusNothing},
		},
	}
	view := briefView(b)
	// The all-quiet statement already says every source was checked.
	if got := view["AlsoChecked"].(string); got != "" {
		t.Errorf("AlsoChecked on an all-quiet brief = %q, want empty", got)
	}
	if stories := view["Stories"].([]storyView); len(stories) != 0 {
		t.Errorf("an all-quiet brief has %d stories, want 0", len(stories))
	}
}

func TestBriefViewItemCap(t *testing.T) {
	var items []contract.Observation
	for i := 0; i < sectionItemCap+3; i++ {
		items = append(items, contract.Observation{Title: fmt.Sprintf("item %d", i)})
	}
	b := engine.Brief{
		Generated: time.Now(),
		Sections: []engine.BriefSection{
			{PluginName: "Busy", Status: contract.StatusOK, Items: items},
		},
	}
	stories := briefView(b)["Stories"].([]storyView)
	if len(stories) != 1 {
		t.Fatalf("want 1 story, got %d", len(stories))
	}
	s := stories[0]
	if len(s.Lead) != sectionItemCap || len(s.Rest) != 3 {
		t.Errorf("fold = %d lead + %d rest, want %d + 3", len(s.Lead), len(s.Rest), sectionItemCap)
	}
	// The fold is presentation only: the section itself stays complete.
	if len(s.Items) != sectionItemCap+3 {
		t.Errorf("story's underlying section lost items: %d", len(s.Items))
	}
	// At or under the cap, nothing folds.
	b.Sections[0].Items = items[:sectionItemCap]
	s = briefView(b)["Stories"].([]storyView)[0]
	if len(s.Lead) != sectionItemCap || len(s.Rest) != 0 {
		t.Errorf("capped-at-limit section folded: %d lead + %d rest", len(s.Lead), len(s.Rest))
	}
}
