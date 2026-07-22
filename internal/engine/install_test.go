package engine

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"orven/contract"
)

const testManifest = `schema_version: 1
id: catalog-plugin
name: Catalog Plugin
version: 1.2.0
publisher: someone
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
collection:
  freshness: 1h
permissions: ["none"]
`

// testCatalogServer serves an index.json and a GitHub-shaped tarball
// (one leading directory component) for a single test plugin.
func testCatalogServer(t *testing.T, extraTarEntries map[string]string) *httptest.Server {
	t.Helper()
	files := map[string]string{
		"repo-main/plugins/community/catalog-plugin/plugin.yaml": testManifest,
		"repo-main/plugins/community/catalog-plugin/main.py":     "print('hi')\n",
		"repo-main/plugins/community/catalog-plugin/README.md":   "test\n",
		"repo-main/unrelated/other.txt":                          "not extracted\n",
	}
	for k, v := range extraTarEntries {
		files[k] = v
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg})
		tw.Write([]byte(content))
	}
	tw.Close()
	gz.Close()
	archive := buf.Bytes()

	index, _ := json.Marshal(CatalogIndex{
		Name: "Test Catalog",
		Plugins: []CatalogPlugin{{
			ID: "catalog-plugin", Name: "Catalog Plugin", Publisher: "someone",
			Version: "1.2.0", Status: "community",
			Path: "plugins/community/catalog-plugin", Permissions: []string{"none"},
		}},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.json":
			w.Write(index)
		case "/archive.tar.gz":
			w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func installEngine(t *testing.T) *Engine {
	t.Helper()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return New(store, t.TempDir())
}

func installFrom(t *testing.T, e *Engine, srv *httptest.Server) {
	t.Helper()
	cat, err := e.Catalog(srv.URL, true)
	if err != nil {
		t.Fatal(err)
	}
	entry := cat.Index.Plugins[0]
	staged, err := e.StageInstall(srv.URL, entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.CommitInstall(staged, srv.URL, entry); err != nil {
		t.Fatal(err)
	}
}

func TestInstallFromCatalog(t *testing.T) {
	e := installEngine(t)
	srv := testCatalogServer(t, nil)

	installFrom(t, e, srv)

	p := e.Plugin("catalog-plugin")
	if p == nil {
		t.Fatal("installed plugin not loaded")
	}
	if p.Manifest.Version != "1.2.0" {
		t.Fatalf("wrong manifest: %+v", p.Manifest)
	}
	rec := e.Store.InstallRecord("catalog-plugin")
	if rec == nil || !rec.Managed || rec.Catalog != srv.URL || rec.Status != "community" {
		t.Fatalf("provenance record wrong: %+v", rec)
	}
	// only the plugin's own folder was extracted
	if _, err := os.Stat(filepath.Join(e.PluginsDir, "catalog-plugin", "plugin.yaml")); err != nil {
		t.Fatal("plugin files missing")
	}
	entries, _ := os.ReadDir(e.PluginsDir)
	if len(entries) != 1 {
		t.Fatalf("unexpected extra content in plugins dir: %v", entries)
	}
}

func TestInstallDuplicateIDRefused(t *testing.T) {
	e := installEngine(t)
	srv := testCatalogServer(t, nil)
	installFrom(t, e, srv)

	cat, _ := e.Catalog(srv.URL, false)
	_, err := e.StageInstall(srv.URL, cat.Index.Plugins[0])
	if err == nil || !strings.Contains(err.Error(), "already installed") {
		t.Fatalf("duplicate install must be refused with the incumbent named, got %v", err)
	}
}

// TestUnsafeArchivePathsNeverEscape: entries that try to traverse out
// of the plugin's folder are neutralized — paths are cleaned before
// prefix matching, so traversal entries fall outside the wanted path
// and are never extracted at all.
func TestUnsafeArchivePathsNeverEscape(t *testing.T) {
	e := installEngine(t)
	srv := testCatalogServer(t, map[string]string{
		"repo-main/plugins/community/catalog-plugin/../../../evil.py":                          "bad\n",
		"repo-main/plugins/community/catalog-plugin/sub/../../../../../../../tmp/evil2.py":     "bad\n",
		"/abs/evil3.py":                                                                        "bad\n",
	})
	cat, _ := e.Catalog(srv.URL, true)
	staged, err := e.StageInstall(srv.URL, cat.Index.Plugins[0])
	if err != nil {
		t.Fatal(err)
	}
	defer e.DiscardStaged(staged)
	var found []string
	filepath.Walk(e.Store.Root, func(p string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() && strings.Contains(info.Name(), "evil") {
			found = append(found, p)
		}
		return nil
	})
	if len(found) != 0 {
		t.Fatalf("traversal entries were extracted: %v", found)
	}
	if _, err := os.Stat(filepath.Join(staged, "plugin.yaml")); err != nil {
		t.Fatal("legitimate plugin files must still extract")
	}
}

// TestUninstallPreservesRunsAndBriefings pins the two agreed rules:
// run history survives uninstall unless deliberately deleted, and
// historical briefings are never altered.
func TestUninstallPreservesRunsAndBriefings(t *testing.T) {
	e := installEngine(t)
	srv := testCatalogServer(t, nil)
	installFrom(t, e, srv)
	id := "catalog-plugin"

	e.Store.SavePluginConfig(id, PluginConfig{Enabled: true, Values: map[string]any{}})
	e.Store.SaveSecrets(id, map[string]string{"api_key": "secret-value-123"})
	e.Store.SaveBatch(StoredBatch{
		PluginID: id, PluginName: "Catalog Plugin", Collected: time.Now(),
		Status: contract.StatusOK,
		Items:  []contract.Observation{{Title: "Something happened", Scope: contract.ScopeEvent}},
	})
	e.Store.AppendRun(id, RunRecord{Started: time.Now(), Finished: time.Now(), Status: contract.StatusOK, Summary: "1 item."})

	time.Sleep(20 * time.Millisecond)
	brief, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(e.Store.Root, "briefs", brief.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}

	if err := e.Uninstall(id, true, false); err != nil {
		t.Fatal(err)
	}

	// gone: files, config, secrets, observations, install record
	if _, err := os.Stat(filepath.Join(e.PluginsDir, id)); !os.IsNotExist(err) {
		t.Fatal("plugin folder must be deleted for a managed uninstall")
	}
	if _, err := os.Stat(filepath.Join(e.Store.Root, "config", id+".json")); !os.IsNotExist(err) {
		t.Fatal("configuration must be removed")
	}
	if _, err := os.Stat(filepath.Join(e.Store.Root, "secrets", id+".json")); !os.IsNotExist(err) {
		t.Fatal("credentials must be removed")
	}
	if _, err := os.Stat(filepath.Join(e.Store.Root, "observations", id)); !os.IsNotExist(err) {
		t.Fatal("raw observations must be removed")
	}
	if e.Store.InstallRecord(id) != nil {
		t.Fatal("install record must be removed")
	}
	if e.Plugin(id) != nil {
		t.Fatal("plugin must disappear from the registry")
	}

	// preserved: run history (by default) and the historical briefing
	if runs := e.Store.Runs(id); len(runs) != 1 {
		t.Fatalf("run history must be preserved by default, got %d records", len(runs))
	}
	after, err := os.ReadFile(filepath.Join(e.Store.Root, "briefs", brief.ID+".json"))
	if err != nil {
		t.Fatal("historical briefing missing after uninstall")
	}
	if !bytes.Equal(before, after) {
		t.Fatal("uninstall altered a historical briefing")
	}
	if got, err := e.Store.Brief(brief.ID); err != nil || len(got.Sections) != 1 || got.Sections[0].Items[0].Title != "Something happened" {
		t.Fatalf("archived briefing must still render its content, got %+v (%v)", got, err)
	}
}

func TestUninstallCanDeleteRunsDeliberately(t *testing.T) {
	e := installEngine(t)
	srv := testCatalogServer(t, nil)
	installFrom(t, e, srv)
	e.Store.AppendRun("catalog-plugin", RunRecord{Started: time.Now(), Status: contract.StatusOK})

	if err := e.Uninstall("catalog-plugin", true, true); err != nil {
		t.Fatal(err)
	}
	if runs := e.Store.Runs("catalog-plugin"); len(runs) != 0 {
		t.Fatal("run history must be deleted when deliberately chosen")
	}
}

// TestSeedLifecycle: seeded exactly once per installation; uninstall
// sticks across restarts; restore is deliberate.
func TestSeedLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	pluginsDir := t.TempDir()
	seedDir := t.TempDir()
	os.MkdirAll(filepath.Join(seedDir, "demo"), 0o755)
	os.WriteFile(filepath.Join(seedDir, "demo", "plugin.yaml"), []byte(`schema_version: 1
id: demo
name: Demo
version: 0.1.0
publisher: orven
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
`), 0o644)
	os.WriteFile(filepath.Join(seedDir, "demo", "main.py"), []byte("print('demo')\n"), 0o644)

	newEngine := func() *Engine {
		store, err := NewStore(dataDir)
		if err != nil {
			t.Fatal(err)
		}
		e := New(store, pluginsDir)
		e.SeedDir = seedDir
		e.SeedOnce()
		return e
	}

	// fresh installation: seeded, managed
	e := newEngine()
	if e.Plugin("demo") == nil {
		t.Fatal("fresh installation must be seeded with the demo")
	}
	if rec := e.Store.InstallRecord("demo"); rec == nil || !rec.Managed || rec.Catalog != "bundled" {
		t.Fatalf("seeded demo must be a managed bundled install, got %+v", rec)
	}

	// uninstall, then simulate container restart/update: must not return
	if err := e.Uninstall("demo", true, false); err != nil {
		t.Fatal(err)
	}
	e2 := newEngine()
	if e2.Plugin("demo") != nil {
		t.Fatal("uninstalled demo must not return after a restart")
	}
	if !e2.SeedAvailable("demo") {
		t.Fatal("restore must be available after uninstall")
	}

	// deliberate restore brings it back
	if err := e2.RestoreSeed("demo"); err != nil {
		t.Fatal(err)
	}
	if e2.Plugin("demo") == nil {
		t.Fatal("restore must reinstall the demo")
	}
	if e2.SeedAvailable("demo") {
		t.Fatal("restore must not be offered while installed")
	}
}
