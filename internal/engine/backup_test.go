package engine

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"orven/contract"
)

func backupEngine(t *testing.T) *Engine {
	t.Helper()
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	pluginsDir := t.TempDir()
	pdir := filepath.Join(pluginsDir, "some-plugin")
	os.MkdirAll(pdir, 0o755)
	os.WriteFile(filepath.Join(pdir, "plugin.yaml"), []byte(`schema_version: 1
id: some-plugin
name: Some Plugin
version: 0.0.1
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
`), 0o644)
	e := New(store, pluginsDir)
	// a realistic data directory: settings, config, secrets, a brief
	e.Store.SaveSettings(DefaultSettings())
	e.Store.SavePluginConfig("some-plugin", PluginConfig{Enabled: true, Values: map[string]any{"url": "http://h"}})
	e.Store.SaveSecrets("some-plugin", map[string]string{"api_key": "the-secret-key-9876"})
	e.Store.SaveBatch(StoredBatch{
		PluginID: "some-plugin", PluginName: "Some Plugin", Collected: time.Now(),
		Status: contract.StatusOK,
		Items:  []contract.Observation{{Title: "A thing happened", Scope: contract.ScopeEvent}},
	})
	return e
}

func TestBackupRoundTripWithEncryptedSecrets(t *testing.T) {
	e := backupEngine(t)
	brief, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := e.WriteBackup(&buf, true, "correct horse"); err != nil {
		t.Fatal(err)
	}
	// the secret value must never appear plain anywhere in the archive
	if bytes.Contains(buf.Bytes(), []byte("the-secret-key-9876")) {
		t.Fatal("secret value appears unencrypted in the backup archive")
	}

	zipPath := filepath.Join(t.TempDir(), "b.zip")
	os.WriteFile(zipPath, buf.Bytes(), 0o644)

	m, err := InspectBackup(zipPath)
	if err != nil || !m.SecretsIncluded || m.Briefs != 1 {
		t.Fatalf("manifest wrong: %+v (%v)", m, err)
	}

	// restore into a fresh installation
	fresh, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fe := New(fresh, t.TempDir())

	// wrong passphrase: refused, nothing restored
	if err := fe.RestoreBackup(zipPath, "wrong"); err == nil || !strings.Contains(err.Error(), "passphrase") {
		t.Fatalf("wrong passphrase must be refused, got %v", err)
	}
	if len(fe.Store.Briefs()) != 0 {
		t.Fatal("a refused restore must not write anything")
	}

	if err := fe.RestoreBackup(zipPath, "correct horse"); err != nil {
		t.Fatal(err)
	}
	got, err := fe.Store.Brief(brief.ID)
	if err != nil || len(got.Sections) != 1 {
		t.Fatalf("restored brief wrong: %+v (%v)", got, err)
	}
	if fe.Store.Secrets("some-plugin")["api_key"] != "the-secret-key-9876" {
		t.Fatal("restored credentials wrong")
	}
	if !fe.Store.PluginConfig("some-plugin").Enabled {
		t.Fatal("restored plugin config wrong")
	}
	// the restore wrote a pre-restore safety backup
	if list := fe.ListBackups(); len(list) != 1 || !strings.Contains(list[0].Name, "prerestore") {
		t.Fatalf("expected a pre-restore safety backup, got %+v", list)
	}
}

func TestBackupWithoutSecretsExcludesThem(t *testing.T) {
	e := backupEngine(t)
	var buf bytes.Buffer
	if err := e.WriteBackup(&buf, false, ""); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(buf.Bytes(), []byte("the-secret-key-9876")) {
		t.Fatal("secrets leaked into a no-credentials backup")
	}
	if bytes.Contains(buf.Bytes(), []byte(secretsEncName)) {
		t.Fatal("no-credentials backup should not carry a secrets block")
	}
	if err := e.WriteBackup(&buf, true, ""); err == nil {
		t.Fatal("including credentials without a passphrase must be refused")
	}
}

func TestBackupRetention(t *testing.T) {
	e := backupEngine(t)
	for i := 0; i < 4; i++ {
		if _, err := e.CreateBackupFile(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(1100 * time.Millisecond) // names carry second precision
	}
	e.PruneBackups(2)
	list := e.ListBackups()
	if len(list) != 2 {
		t.Fatalf("retention must keep exactly 2, got %d", len(list))
	}
	if !list[0].Created.After(list[1].Created) {
		t.Fatal("retention must keep the newest backups")
	}
}

func TestRestoreRefusesForeignArchives(t *testing.T) {
	e := backupEngine(t)

	evil := filepath.Join(t.TempDir(), "evil.zip")
	f, _ := os.Create(evil)
	zw := zip.NewWriter(f)
	w, _ := zw.Create(backupManifestName)
	w.Write([]byte(`{"app":"orven","created":"2026-07-22T00:00:00Z"}`))
	w2, _ := zw.Create("../outside.txt")
	w2.Write([]byte("escape"))
	zw.Close()
	f.Close()

	if err := e.RestoreBackup(evil, ""); err == nil || !strings.Contains(err.Error(), "unexpected entry") {
		t.Fatalf("unexpected entries must abort the restore, got %v", err)
	}

	notOurs := filepath.Join(t.TempDir(), "other.zip")
	f2, _ := os.Create(notOurs)
	zw2 := zip.NewWriter(f2)
	w3, _ := zw2.Create("readme.txt")
	w3.Write([]byte("hello"))
	zw2.Close()
	f2.Close()
	if err := e.RestoreBackup(notOurs, ""); err == nil {
		t.Fatal("archives without an Orven manifest must be refused")
	}
}

// TestRestoreReproducesDemoUninstall: restoring a backup reproduces
// the backed-up state for bundled content — if the demo had been
// uninstalled there, the freshly seeded copy on the restoring machine
// is removed; if it was installed there, it stays.
func TestRestoreReproducesDemoUninstall(t *testing.T) {
	seedDir := t.TempDir()
	os.MkdirAll(filepath.Join(seedDir, "demo"), 0o755)
	os.WriteFile(filepath.Join(seedDir, "demo", "plugin.yaml"), []byte(`schema_version: 1
id: demo
name: Demo
version: 0.1.0
entrypoint: ["python", "main.py"]
engine:
  min_contract: 1
`), 0o644)

	freshSeeded := func() *Engine {
		store, err := NewStore(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		e := New(store, t.TempDir())
		e.SeedDir = seedDir
		e.SeedOnce()
		return e
	}

	// Machine A: seeded, then the demo deliberately uninstalled; back up.
	a := freshSeeded()
	if err := a.Uninstall("demo", true, false); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "a.zip")
	f, _ := os.Create(archive)
	if err := a.WriteBackup(f, false, ""); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Machine B: fresh, so seeding already put the demo back. Restoring
	// A's backup must reproduce A's state: demo uninstalled.
	b := freshSeeded()
	if b.Plugin("demo") == nil {
		t.Fatal("precondition: fresh machine is seeded")
	}
	if err := b.RestoreBackup(archive, ""); err != nil {
		t.Fatal(err)
	}
	if b.Plugin("demo") != nil {
		t.Fatal("restore must reproduce the backed-up uninstall of the bundled demo")
	}
	if !b.SeedAvailable("demo") {
		t.Fatal("the deliberate restore path must remain available afterwards")
	}

	// Inverse: a backup with the demo installed restores without removal.
	c := freshSeeded()
	archive2 := filepath.Join(t.TempDir(), "c.zip")
	f2, _ := os.Create(archive2)
	if err := c.WriteBackup(f2, false, ""); err != nil {
		t.Fatal(err)
	}
	f2.Close()
	d := freshSeeded()
	if err := d.RestoreBackup(archive2, ""); err != nil {
		t.Fatal(err)
	}
	if d.Plugin("demo") == nil {
		t.Fatal("restoring a backup with the demo installed must keep it installed")
	}
}

// TestRestoreIsExactReproduction pins the product meaning of Restore:
// "put me back exactly where I was." State created after the backup —
// newer briefings, configuration for other plugins, run history — is
// removed by the restore (preserved only in the safety backup), while
// domains the archive doesn't carry (credentials, here) are untouched.
func TestRestoreIsExactReproduction(t *testing.T) {
	e := backupEngine(t)
	brief, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "b.zip")
	f, _ := os.Create(archive)
	if err := e.WriteBackup(f, false, ""); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// life goes on after the backup: a newer brief, a new plugin's
	// config and runs (brief IDs carry second precision, so cross a
	// second boundary to guarantee a distinct ID)
	time.Sleep(1100 * time.Millisecond)
	e.Store.SaveBatch(StoredBatch{
		PluginID: "some-plugin", PluginName: "Some Plugin", Collected: time.Now(),
		Status: contract.StatusOK,
		Items:  []contract.Observation{{Title: "Something newer", Scope: contract.ScopeEvent}},
	})
	time.Sleep(20 * time.Millisecond)
	newer, err := e.GenerateBrief()
	if err != nil {
		t.Fatal(err)
	}
	e.Store.SavePluginConfig("later-plugin", PluginConfig{Enabled: true, Values: map[string]any{}})
	e.Store.AppendRun("later-plugin", RunRecord{Started: time.Now(), Status: contract.StatusOK})

	if err := e.RestoreBackup(archive, ""); err != nil {
		t.Fatal(err)
	}

	// exactly the backup's state within backed-up domains
	if _, err := e.Store.Brief(brief.ID); err != nil {
		t.Fatal("the backed-up brief must be present")
	}
	if _, err := e.Store.Brief(newer.ID); err == nil {
		t.Fatal("a brief created after the backup must not survive a restore")
	}
	if len(e.Store.Briefs()) != 1 {
		t.Fatalf("expected exactly the backup's 1 brief, got %d", len(e.Store.Briefs()))
	}
	if _, err := os.Stat(filepath.Join(e.Store.Root, "config", "later-plugin.json")); !os.IsNotExist(err) {
		t.Fatal("config created after the backup must not survive")
	}
	if runs := e.Store.Runs("later-plugin"); len(runs) != 0 {
		t.Fatal("run history created after the backup must not survive")
	}
	// a domain the archive doesn't carry stays untouched
	if e.Store.Secrets("some-plugin")["api_key"] != "the-secret-key-9876" {
		t.Fatal("credentials must be untouched when the archive carries none")
	}
	// and the removed state is preserved in the safety backup
	list := e.ListBackups()
	if len(list) != 1 || !strings.Contains(list[0].Name, "prerestore") {
		t.Fatalf("expected the pre-restore safety backup, got %+v", list)
	}
}

func TestClockDue(t *testing.T) {
	base := time.Date(2026, 7, 22, 3, 0, 0, 0, time.Local)
	if !clockDue(base, base.Add(45*time.Minute), "03:30") {
		t.Fatal("03:30 falls in the window and must be due")
	}
	if clockDue(base, base.Add(15*time.Minute), "03:30") {
		t.Fatal("03:30 is after the window and must not be due")
	}
	if clockDue(base.Add(40*time.Minute), base.Add(50*time.Minute), "03:30") {
		t.Fatal("03:30 is before the window and must not be due")
	}
}
