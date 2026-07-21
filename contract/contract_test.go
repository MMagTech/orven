package contract

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestZeroTimesAreAbsentFromJSON pins the contract-freeze rule: unset
// timestamps are absent from the wire, never a year-one placeholder a
// plugin author would see, copy, or code around.
func TestZeroTimesAreAbsentFromJSON(t *testing.T) {
	obs, _ := json.Marshal(Observation{Title: "Backup completed"})
	if strings.Contains(string(obs), "occurred_at") {
		t.Fatalf("zero occurred_at must be omitted, got %s", obs)
	}
	in, _ := json.Marshal(Input{ContractVersion: 1, PluginID: "x", Now: time.Now()})
	if strings.Contains(string(in), "window_start") {
		t.Fatalf("zero window_start must be omitted on first run, got %s", in)
	}

	when := time.Date(2026, 7, 21, 3, 12, 0, 0, time.UTC)
	obs2, _ := json.Marshal(Observation{Title: "Backup completed", OccurredAt: when})
	if !strings.Contains(string(obs2), `"occurred_at":"2026-07-21T03:12:00Z"`) {
		t.Fatalf("set occurred_at must be present, got %s", obs2)
	}
	in2, _ := json.Marshal(Input{ContractVersion: 1, PluginID: "x", Now: time.Now(), WindowStart: when})
	if !strings.Contains(string(in2), `"window_start":"2026-07-21T03:12:00Z"`) {
		t.Fatalf("set window_start must be present, got %s", in2)
	}
}
