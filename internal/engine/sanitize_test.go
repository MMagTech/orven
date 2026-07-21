package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"orven/contract"
)

const testSecret = "sk-live-9f8e7d6c5b4a"

// TestSanitizerRedacts: assigned values and well-defined credential
// contexts are always redacted.
func TestSanitizerRedacts(t *testing.T) {
	san := newSanitizer(map[string]string{"api_key": testSecret})
	cases := []struct{ in, want string }{
		{"the key is " + testSecret + " ok", "the key is [redacted] ok"},
		{testSecret, "[redacted]"},
		{"GET http://radarr:7878/api?apikey=abc123&pageSize=50", "GET http://radarr:7878/api?apikey=[redacted]&pageSize=50"},
		{"url was http://h/x?api_key=zz9&t=1", "url was http://h/x?api_key=[redacted]&t=1"},
		{"retried with &token=deadbeef", "retried with &token=[redacted]"},
		{"posted ?password=hunter2 to the form", "posted ?password=[redacted] to the form"},
		{"header Authorization: Bearer eyJhbGciOi was sent", "header Authorization: Bearer [redacted] was sent"},
		{"sent authorization: abc123", "sent authorization: [redacted]"},
		{"X-Api-Key: abc123 rejected", "X-Api-Key: [redacted] rejected"},
		{"Api-Key: qrs987", "Api-Key: [redacted]"},
	}
	for _, c := range cases {
		if got := san.clean(c.in); got != c.want {
			t.Errorf("clean(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestSanitizerIsNonDestructive: legitimate plugin text with no
// credential present passes through byte-for-byte.
func TestSanitizerIsNonDestructive(t *testing.T) {
	san := newSanitizer(map[string]string{"api_key": testSecret, "tiny": "ab"})
	benign := []string{
		"2 tokens expired on the auth server",
		"The API key was rejected.",
		"Password change detected on the admin account",
		"GET http://radarr:7878/api/v3/queue?pageSize=100 answered in 40 ms",
		"Authorization failed for the request",
		"3 movies finished downloading — Dune: Part Two is ready to watch.",
		"The author=smith filter matched 4 items",
		"backup of secrets folder completed, 3 files",
		"", // empty stays empty
	}
	for _, s := range benign {
		if got := san.clean(s); got != s {
			t.Errorf("benign text was altered:\n  in:  %q\n  out: %q", s, got)
		}
	}
}

// writeSecretPlugin creates a plugin whose behavior is driven by its
// mode config: echo (secret in stdout fields), crash (secret in a
// stderr traceback), clean (benign output only).
func writeSecretPlugin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	pdir := filepath.Join(dir, "leaky")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `schema_version: 1
id: leaky
name: Leaky
version: 0.0.1
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
config:
  - key: mode
    type: text
    label: Mode
    default: clean
  - key: api_key
    type: secret
    label: Key
`
	script := `import json, sys
inp = json.load(sys.stdin)
mode = inp["config"].get("mode", "clean")
key = inp.get("secrets", {}).get("api_key", "")
if mode == "crash":
    sys.stderr.write("Traceback: HTTPError for http://h/api?apikey=%s auth %s\n" % (key, key))
    sys.exit(3)
if mode == "echo":
    json.dump({"contract_version": 1, "status": "ok",
        "summary": "Checked http://h/api?apikey=%s fine" % key,
        "observations": [
            {"title": "Key is " + key, "body": "Sent Authorization: Bearer " + key, "scope": "state"},
        ]}, sys.stdout)
else:
    json.dump({"contract_version": 1, "status": "ok",
        "summary": "The API key was rejected once, then accepted.",
        "observations": [
            {"title": "2 tokens expired on the auth server",
             "body": "GET http://radarr:7878/api/v3/queue?pageSize=100 answered normally.",
             "scope": "event"},
        ]}, sys.stdout)
`
	os.WriteFile(filepath.Join(pdir, "plugin.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(pdir, "main.py"), []byte(script), 0o644)
	return dir
}

func secretEngine(t *testing.T, mode string) (*Engine, *Plugin) {
	t.Helper()
	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not on PATH")
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := New(store, writeSecretPlugin(t))
	p := e.Plugin("leaky")
	if p == nil {
		t.Fatal("leaky plugin not loaded")
	}
	e.Store.SavePluginConfig("leaky", PluginConfig{Enabled: true, Values: map[string]any{"mode": mode}})
	e.Store.SaveSecrets("leaky", map[string]string{"api_key": testSecret})
	return e, p
}

func assertNoSecret(t *testing.T, where, s string) {
	t.Helper()
	if strings.Contains(s, testSecret) {
		t.Fatalf("credential persisted in %s: %q", where, s)
	}
}

// TestCredentialCannotBePersistedFromStdout: a plugin that echoes its
// assigned secret in summary, title, and body must have it redacted in
// the stored batch, run history, engine log, and the compiled brief.
func TestCredentialCannotBePersistedFromStdout(t *testing.T) {
	e, p := secretEngine(t, "echo")
	batch, err := e.TryRun(p, true)
	if err != nil {
		t.Fatal(err)
	}
	assertNoSecret(t, "returned batch summary", batch.Summary)
	for _, o := range batch.Items {
		assertNoSecret(t, "observation title", o.Title)
		assertNoSecret(t, "observation body", o.Body)
	}
	if !strings.Contains(batch.Items[0].Title, redacted) {
		t.Fatalf("expected redaction marker in echoed title, got %q", batch.Items[0].Title)
	}
	// query-param pattern also caught, independent of the exact value
	if !strings.Contains(batch.Summary, "apikey="+redacted) {
		t.Fatalf("credential query parameter not redacted: %q", batch.Summary)
	}

	// on disk
	blob := readAll(t, filepath.Join(e.Store.Root, "observations", "leaky"))
	assertNoSecret(t, "stored observation files", blob)
	runs := readAll(t, filepath.Join(e.Store.Root, "runs"))
	assertNoSecret(t, "run history", runs)
	assertNoSecret(t, "engine log", strings.Join(e.LogLines(), "\n"))

	brief, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	bj, _ := json.Marshal(brief)
	assertNoSecret(t, "compiled brief", string(bj))
}

// TestCredentialCannotBePersistedFromStderr: a crash that prints the
// secret to stderr must not carry it into the run record or log.
func TestCredentialCannotBePersistedFromStderr(t *testing.T) {
	e, p := secretEngine(t, "crash")
	if _, err := e.TryRun(p, true); err != nil {
		t.Fatal(err)
	}
	attempt, _ := e.Store.LastRun("leaky")
	if attempt == nil || attempt.Status != contract.StatusError {
		t.Fatalf("expected an error run record, got %+v", attempt)
	}
	assertNoSecret(t, "run record error", attempt.Error)
	if !strings.Contains(attempt.Error, redacted) {
		t.Fatalf("stderr-derived error should show redaction, got %q", attempt.Error)
	}
	assertNoSecret(t, "engine log", strings.Join(e.LogLines(), "\n"))
	assertNoSecret(t, "run history on disk", readAll(t, filepath.Join(e.Store.Root, "runs")))
}

// TestWellBehavedOutputIsUntouched: with a secret assigned but never
// exposed, every user-visible field survives byte-for-byte — including
// prose about tokens, keys, and URLs with benign query parameters.
func TestWellBehavedOutputIsUntouched(t *testing.T) {
	e, p := secretEngine(t, "clean")
	batch, err := e.TryRun(p, true)
	if err != nil {
		t.Fatal(err)
	}
	if batch.Summary != "The API key was rejected once, then accepted." {
		t.Fatalf("benign summary altered: %q", batch.Summary)
	}
	if batch.Items[0].Title != "2 tokens expired on the auth server" {
		t.Fatalf("benign title altered: %q", batch.Items[0].Title)
	}
	if batch.Items[0].Body != "GET http://radarr:7878/api/v3/queue?pageSize=100 answered normally." {
		t.Fatalf("benign body altered: %q", batch.Items[0].Body)
	}
}

func readAll(t *testing.T, dir string) string {
	t.Helper()
	var sb strings.Builder
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			b, _ := os.ReadFile(path)
			fmt.Fprintf(&sb, "%s\n", b)
		}
		return nil
	})
	return sb.String()
}
