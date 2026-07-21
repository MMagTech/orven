package validate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTitleStyle(t *testing.T) {
	cases := []struct {
		title      string
		wantWarn   bool
		suggestion string // "" = no suggestion expected
	}{
		{"1 episode is stuck in the queue", false, ""},
		{"3 movies finished downloading", false, ""},
		{"Certificate renewed", false, ""},
		{"2 episodes of Slow Horses downloaded", false, ""}, // proper noun minority
		{"418 GB backed up on RAID6", false, ""},            // acronyms don't count
		{"2 New Requests Awaiting Approval", true, "2 new requests awaiting approval   (verify proper nouns)"},
		{"1 Episode Stuck in Queue", true, "1 episode stuck in queue   (verify proper nouns)"},
		{"3 Drives Failed on RAID Array", true, "3 drives failed on RAID array   (verify proper nouns)"},
		{"Backup completed.", true, "Backup completed"},
		{"Backup failed!", true, "Backup failed"},
		{"RAID ARRAY DEGRADED", true, ""}, // all-caps: warn, no suggestion
		{"The overnight backup completed successfully after the second attempt at three in the morning", true, ""},
	}
	for _, c := range cases {
		r := &report{}
		titleStyle(r, c.title)
		if got := len(r.findings) > 0; got != c.wantWarn {
			t.Errorf("%q: warned=%v, want %v (%+v)", c.title, got, c.wantWarn, r.findings)
			continue
		}
		if c.suggestion != "" {
			if len(r.findings) == 0 || r.findings[0].Suggestion != c.suggestion {
				t.Errorf("%q: suggestion = %q, want %q", c.title, first(r.findings), c.suggestion)
			}
		}
		// Hard boundary: any suggestion differs by capitalization or
		// trailing punctuation only — same words, same order.
		for _, f := range r.findings {
			if f.Suggestion == "" {
				continue
			}
			sug := strings.TrimSuffix(f.Suggestion, "   (verify proper nouns)")
			if !strings.EqualFold(sug, strings.TrimRight(c.title, ".!")) {
				t.Errorf("%q: suggestion %q changed more than capitalization/punctuation", c.title, sug)
			}
		}
	}
}

func first(fs []Finding) string {
	if len(fs) == 0 {
		return ""
	}
	return fs[0].Suggestion
}

func TestDemoPluginValidatesClean(t *testing.T) {
	needPython(t)
	findings := Dir(repoPath(t, "plugins", "demo-activity"))
	if len(findings) != 0 {
		t.Fatalf("the reference plugin must validate clean, got %+v", findings)
	}
}

func TestHTTPExampleValidatesClean(t *testing.T) {
	needPython(t)
	findings := Dir(repoPath(t, "examples", "radarr-queue"))
	if len(findings) != 0 {
		t.Fatalf("the HTTP example plugin must validate clean, got %+v", findings)
	}
}

func TestBrokenManifestIsAnError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte("schema_version: 1\nname: No Entrypoint\n"), 0o644)
	if !hasError(Dir(dir), "id and entrypoint") {
		t.Fatal("missing id/entrypoint must be an error")
	}
}

func TestMisbehavingOutputIsCaught(t *testing.T) {
	needPython(t)
	// A plugin that leaks its secret, uses forbidden voice, reports an
	// unknown status, and forgets contract_version.
	dir := writePlugin(t, `import json, sys
inp = json.load(sys.stdin)
json.dump({
    "status": "weird",
    "summary": "You should restart the server. Also the key is " + inp["secrets"]["api_key"] + ".",
    "observations": [{"title": "Server Needs A Restart.", "scope": "sideways"}],
}, sys.stdout)
`)
	findings := Dir(dir)
	for _, want := range []string{"contract_version", "status \"weird\"", "forbidden voice", "secret leakage", "unknown scope"} {
		if !hasError(findings, want) {
			t.Errorf("expected an error mentioning %q, got %+v", want, findings)
		}
	}
}

func TestUnknownKindIsAWarning(t *testing.T) {
	needPython(t)
	dir := writePlugin(t, `import json, sys
json.load(sys.stdin)
json.dump({"contract_version": 1, "status": "ok", "summary": "Checked.",
    "observations": [{"title": "Queue checked", "kind": "urgent"}]}, sys.stdout)
`)
	found := false
	for _, f := range Dir(dir) {
		if f.Severity == "WARN" && strings.Contains(f.Message, `unknown kind "urgent"`) {
			found = true
		}
	}
	if !found {
		t.Fatal("unknown kind must draw a warning")
	}
}

func TestCredentialShapedOutputIsAnError(t *testing.T) {
	needPython(t)
	dir := writePlugin(t, `import json, sys
json.load(sys.stdin)
json.dump({"contract_version": 1, "status": "ok",
    "summary": "Checked the queue.",
    "observations": [{"title": "Queue checked",
        "body": "Called http://radarr:7878/api?apikey=abc123 with Authorization: Bearer xyz."}],
}, sys.stdout)
`)
	if !hasError(Dir(dir), "credential-shaped content") {
		t.Fatal("credential-shaped output must be an error")
	}
}

func TestCredentialShapedFixtureIsAWarning(t *testing.T) {
	needPython(t)
	dir := writePlugin(t, `import json, sys
json.load(sys.stdin)
json.dump({"contract_version": 1, "status": "nothing", "summary": "Nothing new."}, sys.stdout)
`)
	os.WriteFile(filepath.Join(dir, "fixtures", "f.json"),
		[]byte(`{"url": "http://h/api?api_key=real-key-oops"}`), 0o644)
	found := false
	for _, f := range Dir(dir) {
		if f.Severity == "WARN" && strings.Contains(f.Where, "fixtures/") &&
			strings.Contains(f.Message, "credential-shaped") {
			found = true
		}
	}
	if !found {
		t.Fatal("credential-shaped fixture content must be warned about")
	}
}

func TestTimeoutAndIntervalErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(`schema_version: 1
id: bad-durations
name: Bad Durations
version: 0.0.1
entrypoint: ["python", "main.py"]
collection:
  recommended_interval: soonish
  min_interval: 4h
  max_interval: 1h
`), 0o644)
	findings := Dir(dir)
	if !hasError(findings, "not a valid duration") {
		t.Errorf("unparsable duration must be an error: %+v", findings)
	}
	if !hasError(findings, "greater than max_interval") {
		t.Errorf("min > max must be an error: %+v", findings)
	}
}

// ---- helpers ----

func needPython(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("python"); err != nil {
		t.Skip("python not on PATH")
	}
}

func repoPath(t *testing.T, parts ...string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(append([]string{root}, parts...)...)
}

func writePlugin(t *testing.T, mainPy string) string {
	t.Helper()
	dir := t.TempDir()
	manifest := `schema_version: 1
id: misbehaving
name: Misbehaving
version: 0.0.1
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
collection:
  freshness: 1h
permissions: ["none"]
config:
  - key: api_key
    type: secret
    label: API key
    required: true
`
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(dir, "main.py"), []byte(mainPy), 0o644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("test plugin"), 0o644)
	os.MkdirAll(filepath.Join(dir, "fixtures"), 0o755)
	os.WriteFile(filepath.Join(dir, "fixtures", "f.json"), []byte("{}"), 0o644)
	os.MkdirAll(filepath.Join(dir, "tests"), 0o755)
	os.WriteFile(filepath.Join(dir, "tests", "t.py"), []byte("# placeholder"), 0o644)
	return dir
}

func hasError(fs []Finding, substr string) bool {
	for _, f := range fs {
		if f.Severity == "ERROR" && strings.Contains(f.Message, substr) {
			return true
		}
	}
	return false
}
