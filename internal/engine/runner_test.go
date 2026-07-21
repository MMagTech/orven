package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"orven/contract"
)

// realPythonPath finds an actual working Python interpreter to stand
// behind the test shim, skipping Windows Store alias stubs that exist
// on PATH but only print an install hint.
func realPythonPath(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"python", "python3"} {
		p, err := exec.LookPath(name)
		if err != nil || strings.Contains(p, "WindowsApps") {
			continue
		}
		if abs, err := filepath.Abs(p); err == nil {
			return abs
		}
	}
	t.Skip("no working python interpreter found")
	return ""
}

// TestResolveInterpreter pins the fallback's exact boundaries: only the
// bare names "python" and "python3" are ever substituted, only when the
// declared one is missing and the sibling exists, and every other
// entrypoint — paths, other interpreters, near-misses — passes through
// untouched no matter what is or isn't installed.
func TestResolveInterpreter(t *testing.T) {
	have := func(names ...string) func(string) (string, error) {
		set := map[string]bool{}
		for _, n := range names {
			set[n] = true
		}
		return func(n string) (string, error) {
			if set[n] {
				return "/fake/bin/" + n, nil
			}
			return "", exec.ErrNotFound
		}
	}

	cases := []struct {
		name      string
		declared  string
		installed []string
		want      string
	}{
		{"declared python present", "python", []string{"python"}, "python"},
		{"python missing, python3 present", "python", []string{"python3"}, "python3"},
		{"python3 missing, python present", "python3", []string{"python"}, "python"},
		{"declared python3 present, sibling ignored", "python3", []string{"python", "python3"}, "python3"},
		{"neither present: declared name kept", "python", nil, "python"},
		// Everything below must never be altered, whatever is installed.
		{"absolute path untouched", "/usr/bin/python", []string{"python3"}, "/usr/bin/python"},
		{"python2 untouched", "python2", []string{"python", "python3"}, "python2"},
		{"python.exe untouched", "python.exe", []string{"python3"}, "python.exe"},
		{"node untouched even when missing", "node", []string{"python", "python3", "nodejs"}, "node"},
		{"arbitrary binary untouched", "my-plugin-binary", nil, "my-plugin-binary"},
	}
	for _, c := range cases {
		if got := resolveInterpreter(c.declared, have(c.installed...)); got != c.want {
			t.Errorf("%s: resolveInterpreter(%q) = %q, want %q", c.name, c.declared, got, c.want)
		}
	}
}

// TestInterpreterFallbackEndToEnd simulates a system that ships only
// `python3` (stock Debian/Ubuntu): PATH is restricted to a shim
// directory whose sole entry is a working python3. A plugin whose
// manifest declares `python` must still run successfully through the
// engine's fallback.
func TestInterpreterFallbackEndToEnd(t *testing.T) {
	real := realPythonPath(t)

	shimDir := t.TempDir()
	if runtime.GOOS == "windows" {
		shim := "@echo off\r\n\"" + real + "\" %*\r\n"
		if err := os.WriteFile(filepath.Join(shimDir, "python3.bat"), []byte(shim), 0o755); err != nil {
			t.Fatal(err)
		}
	} else {
		shim := "#!/bin/sh\nexec \"" + real + "\" \"$@\"\n"
		if err := os.WriteFile(filepath.Join(shimDir, "python3"), []byte(shim), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", shimDir) // now `python` does not exist; only `python3` does

	const declare = "python"
	dir := t.TempDir()
	pdir := filepath.Join(dir, "fallback-check")
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf(`schema_version: 1
id: fallback-check
name: Fallback Check
version: 0.0.1
entrypoint: ["%s", "main.py"]
engine:
  min_contract: 1
timeout: 30s
`, declare)
	script := `import json, sys
json.load(sys.stdin)
json.dump({"contract_version": 1, "status": "nothing", "summary": "ran"}, sys.stdout)
`
	if err := os.WriteFile(filepath.Join(pdir, "plugin.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdir, "main.py"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := New(store, dir)
	p := e.Plugin("fallback-check")
	if p == nil {
		t.Fatal("fallback-check plugin not loaded")
	}
	batch, err := e.TryRun(p, true)
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != contract.StatusNothing {
		t.Fatalf("plugin declaring the missing interpreter name must still run, got %q", batch.Status)
	}
}
