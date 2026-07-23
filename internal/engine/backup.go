package engine

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Backups are user-facing (Daily Care): a backup is one zip of the
// data directory a person can download, schedule, browse, and restore.
// Credentials are only ever included encrypted under a passphrase
// (CONSTRAINTS §13: no unencrypted secret backups) — everything else
// in a backup is the same plain files the data directory holds.
//
// A backup covers the data directory only. Plugin folders are code,
// reinstallable from catalogs; the install records inside the backup
// say what was installed and from where.

// backupRoots are the only entries a backup contains and the only
// entries a restore will accept.
var backupRoots = []string{"briefs", "observations", "config", "runs", "installed"}
var backupFiles = []string{"settings.json", "seeded", "onboarding"}

const (
	backupManifestName = "backup-manifest.json"
	secretsEncName     = "secrets.enc"
	backupPassFile     = "backup-passphrase"
	pbkdf2Iterations   = 600_000
)

type BackupManifest struct {
	App             string    `json:"app"`
	Created         time.Time `json:"created"`
	SecretsIncluded bool      `json:"secrets_included"`
	Briefs          int       `json:"briefs"`
	Plugins         int       `json:"configured_plugins"`
}

type BackupInfo struct {
	Name            string
	Size            int64
	Created         time.Time
	SecretsIncluded bool
}

// ---- backup settings (part of app settings) ----

type BackupSettings struct {
	Enabled        bool   `json:"enabled"`
	Time           string `json:"time"`            // "03:30"
	Dir            string `json:"dir"`             // "" = <data>/backups
	Retention      int    `json:"retention"`       // how many automatic backups to keep
	IncludeSecrets bool   `json:"include_secrets"` // requires a stored passphrase
}

// BackupDir resolves the configured backup folder.
func (e *Engine) BackupDir() string {
	d := e.Store.Settings().Backup.Dir
	if strings.TrimSpace(d) == "" {
		return filepath.Join(e.Store.Root, "backups")
	}
	return d
}

// The automatic-backup passphrase is write-only via the UI, stored so
// scheduled backups can include credentials. It protects the backup
// copies wherever they travel; the live data directory is unchanged.
func (s *Store) BackupPassphrase() string {
	b, err := os.ReadFile(filepath.Join(s.Root, backupPassFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func (s *Store) SaveBackupPassphrase(p string) error {
	return os.WriteFile(filepath.Join(s.Root, backupPassFile), []byte(p+"\n"), 0o600)
}

// ---- writing backups ----

// WriteBackup streams one backup zip. includeSecrets requires a
// non-empty passphrase; credentials are AES-256-GCM encrypted with a
// PBKDF2-derived key and never stored plain in any archive.
func (e *Engine) WriteBackup(w io.Writer, includeSecrets bool, passphrase string) error {
	if includeSecrets && strings.TrimSpace(passphrase) == "" {
		return fmt.Errorf("including credentials requires a passphrase")
	}
	zw := zip.NewWriter(w)

	manifest := BackupManifest{
		App: "orven", Created: time.Now().UTC().Truncate(time.Second),
		SecretsIncluded: includeSecrets,
	}
	manifest.Briefs = len(e.Store.Briefs())
	if entries, err := os.ReadDir(filepath.Join(e.Store.Root, "config")); err == nil {
		manifest.Plugins = len(entries)
	}
	mw, err := zw.Create(backupManifestName)
	if err != nil {
		return err
	}
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	mw.Write(mb)

	addFile := func(rel, abs string) error {
		info, err := os.Stat(abs)
		if err != nil || info.IsDir() {
			return nil
		}
		f, err := os.Open(abs)
		if err != nil {
			return err
		}
		defer f.Close()
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		return err
	}
	for _, root := range backupRoots {
		base := filepath.Join(e.Store.Root, root)
		filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(e.Store.Root, p)
			if err != nil {
				return nil
			}
			return addFile(filepath.ToSlash(rel), p)
		})
	}
	for _, f := range backupFiles {
		if err := addFile(f, filepath.Join(e.Store.Root, f)); err != nil {
			return err
		}
	}

	if includeSecrets {
		all := map[string]map[string]string{}
		entries, _ := os.ReadDir(filepath.Join(e.Store.Root, "secrets"))
		for _, ent := range entries {
			id := strings.TrimSuffix(ent.Name(), ".json")
			if m := e.Store.Secrets(id); len(m) > 0 {
				all[id] = m
			}
		}
		plain, _ := json.Marshal(all)
		enc, err := encryptWithPassphrase(plain, passphrase)
		if err != nil {
			return err
		}
		w, err := zw.Create(secretsEncName)
		if err != nil {
			return err
		}
		w.Write(enc)
	}
	return zw.Close()
}

// CreateBackupFile writes a backup into the backup folder using the
// configured automatic-backup options; returns the file name.
func (e *Engine) CreateBackupFile() (string, error) {
	cfg := e.Store.Settings().Backup
	includeSecrets := cfg.IncludeSecrets && e.Store.BackupPassphrase() != ""
	dir := e.BackupDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := "orven-backup-" + time.Now().Format("20060102-150405") + ".zip"
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	if err := e.WriteBackup(f, includeSecrets, e.Store.BackupPassphrase()); err != nil {
		f.Close()
		os.Remove(filepath.Join(dir, name))
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	e.Logf("backup written: %s (credentials included: %v)", name, includeSecrets)
	return name, nil
}

// ListBackups reads the backup folder, newest first.
func (e *Engine) ListBackups() []BackupInfo {
	var out []BackupInfo
	entries, _ := os.ReadDir(e.BackupDir())
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".zip") {
			continue
		}
		info, err := ent.Info()
		if err != nil {
			continue
		}
		b := BackupInfo{Name: ent.Name(), Size: info.Size(), Created: info.ModTime()}
		if m, err := readBackupManifest(filepath.Join(e.BackupDir(), ent.Name())); err == nil {
			b.Created = m.Created
			b.SecretsIncluded = m.SecretsIncluded
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out
}

// PruneBackups keeps the newest `keep` backups in the backup folder.
func (e *Engine) PruneBackups(keep int) {
	if keep <= 0 {
		return
	}
	list := e.ListBackups()
	for i := keep; i < len(list); i++ {
		os.Remove(filepath.Join(e.BackupDir(), list[i].Name))
		e.Logf("backup pruned by retention: %s", list[i].Name)
	}
}

func readBackupManifest(zipPath string) (BackupManifest, error) {
	var m BackupManifest
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return m, fmt.Errorf("not a readable backup archive")
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name == backupManifestName {
			r, err := f.Open()
			if err != nil {
				return m, err
			}
			defer r.Close()
			if err := json.NewDecoder(r).Decode(&m); err != nil {
				return m, fmt.Errorf("backup manifest is malformed")
			}
			if m.App != "orven" {
				return m, fmt.Errorf("this archive was not created by Orven")
			}
			return m, nil
		}
	}
	return m, fmt.Errorf("this archive has no Orven backup manifest")
}

// InspectBackup returns the manifest of a backup archive on disk.
func InspectBackup(zipPath string) (BackupManifest, error) {
	return readBackupManifest(zipPath)
}

// ---- restore ----

// RestoreBackup applies a backup archive to the data directory,
// implementing the product meaning of Restore (decided 2026-07-22):
// "put me back exactly where I was." Within the backed-up domains the
// result is a strict reproduction of the archive — anything created
// since the backup is removed (the mandatory pre-restore safety
// backup preserves it). Domains the archive does not carry are left
// alone: credentials stay as they are unless the backup includes
// them, and plugin folders are never touched (they are not data).
// Everything is validated — including decrypting credentials with the
// given passphrase — before anything is changed.
func (e *Engine) RestoreBackup(zipPath, passphrase string) error {
	manifest, err := readBackupManifest(zipPath)
	if err != nil {
		return err
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("not a readable backup archive")
	}
	defer zr.Close()

	// validate every entry and pre-read secrets before changing anything
	var secrets map[string]map[string]string
	archiveSeeded := false
	archiveRecords := map[string]bool{} // install records present in the archive
	for _, f := range zr.File {
		name := path.Clean(f.Name)
		if name == "seeded" {
			archiveSeeded = true
		}
		if strings.HasPrefix(name, "installed/") && strings.HasSuffix(name, ".json") {
			archiveRecords[strings.TrimSuffix(path.Base(name), ".json")] = true
		}
		switch {
		case name == backupManifestName:
		case name == secretsEncName:
			if !manifest.SecretsIncluded {
				return fmt.Errorf("archive is inconsistent about credentials")
			}
			r, err := f.Open()
			if err != nil {
				return err
			}
			enc, err := io.ReadAll(io.LimitReader(r, 8<<20))
			r.Close()
			if err != nil {
				return err
			}
			plain, err := decryptWithPassphrase(enc, passphrase)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(plain, &secrets); err != nil {
				return fmt.Errorf("decrypted credentials are malformed")
			}
		case isAllowedBackupEntry(name):
		default:
			return fmt.Errorf("archive contains an unexpected entry (%s) — refusing to restore", name)
		}
	}
	if manifest.SecretsIncluded && secrets == nil {
		return fmt.Errorf("the archive says credentials are included but none were found")
	}

	// Snapshot which seed plugins are bundled installs before the
	// records are cleared — the reconciliation below needs this fact
	// after the strict wipe has removed the records that prove it.
	preBundled := map[string]bool{}
	if e.SeedDir != "" {
		if entries, err := os.ReadDir(e.SeedDir); err == nil {
			for _, ent := range entries {
				if rec := e.Store.InstallRecord(ent.Name()); rec != nil && rec.Catalog == "bundled" {
					preBundled[ent.Name()] = true
				}
			}
		}
	}

	// safety backup of the current state, then strict reproduction:
	// clear the backed-up domains so nothing newer than the archive
	// survives inside them, then apply the archive's contents.
	if _, err := e.CreateSafetyBackup(); err != nil {
		return fmt.Errorf("could not write the pre-restore safety backup: %v", err)
	}
	for _, root := range backupRoots {
		os.RemoveAll(filepath.Join(e.Store.Root, root))
	}
	for _, f := range backupFiles {
		os.Remove(filepath.Join(e.Store.Root, f))
	}
	if secrets != nil {
		// the archive carries credentials: reproduce that domain too
		os.RemoveAll(filepath.Join(e.Store.Root, "secrets"))
		os.MkdirAll(filepath.Join(e.Store.Root, "secrets"), 0o755)
	}
	for _, f := range zr.File {
		name := path.Clean(f.Name)
		if name == backupManifestName || name == secretsEncName || !isAllowedBackupEntry(name) {
			continue
		}
		target := filepath.Join(e.Store.Root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			r.Close()
			return err
		}
		_, err = io.Copy(out, io.LimitReader(r, 64<<20))
		r.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	for id, m := range secrets {
		if strings.ContainsAny(id, `/\.`) {
			continue
		}
		e.Store.SaveSecrets(id, m)
	}
	e.Reload()

	// Seed reconciliation: restore reproduces the backed-up state for
	// bundled content too. If the backed-up installation had been
	// seeded and held no install record for a seed plugin, that plugin
	// had been uninstalled there — so the copy this (freshly seeded)
	// installation carries is removed as part of the restore.
	if e.SeedDir != "" && archiveSeeded {
		for id := range preBundled {
			if archiveRecords[id] || e.Plugin(id) == nil {
				continue
			}
			if err := e.Uninstall(id, true, false); err != nil {
				e.Logf("restore: could not remove bundled plugin %s: %v", id, err)
			} else {
				e.Logf("restore: bundled plugin %s removed to match the backup", id)
			}
		}
	}

	e.Logf("backup restored from %s (created %s)", filepath.Base(zipPath), manifest.Created.Format("2006-01-02 15:04"))
	return nil
}

// CreateSafetyBackup writes an unencrypted-secrets-free backup of the
// current state into the backup folder before a restore overwrites it.
func (e *Engine) CreateSafetyBackup() (string, error) {
	dir := e.BackupDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := "orven-prerestore-" + time.Now().Format("20060102-150405") + ".zip"
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := e.WriteBackup(f, false, ""); err != nil {
		os.Remove(filepath.Join(dir, name))
		return "", err
	}
	return name, nil
}

func isAllowedBackupEntry(name string) bool {
	if strings.Contains(name, "..") || path.IsAbs(name) || strings.Contains(name, "\\") {
		return false
	}
	for _, f := range backupFiles {
		if name == f {
			return true
		}
	}
	for _, root := range backupRoots {
		if strings.HasPrefix(name, root+"/") {
			return true
		}
	}
	return false
}

// ---- passphrase encryption (AES-256-GCM, PBKDF2-SHA256 key) ----

type encEnvelope struct {
	KDF   string `json:"kdf"`
	Iter  int    `json:"iter"`
	Salt  string `json:"salt"`
	Nonce string `json:"nonce"`
	Data  string `json:"data"`
}

func encryptWithPassphrase(plain []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	key, err := pbkdf2.Key(sha256.New, passphrase, salt, pbkdf2Iterations, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plain, nil)
	return json.Marshal(encEnvelope{
		KDF: "pbkdf2-sha256", Iter: pbkdf2Iterations,
		Salt:  base64.StdEncoding.EncodeToString(salt),
		Nonce: base64.StdEncoding.EncodeToString(nonce),
		Data:  base64.StdEncoding.EncodeToString(sealed),
	})
}

func decryptWithPassphrase(enc []byte, passphrase string) ([]byte, error) {
	if strings.TrimSpace(passphrase) == "" {
		return nil, fmt.Errorf("this backup's credentials are encrypted — the passphrase is required")
	}
	var env encEnvelope
	if err := json.Unmarshal(enc, &env); err != nil || env.KDF != "pbkdf2-sha256" {
		return nil, fmt.Errorf("the encrypted credentials block is malformed")
	}
	salt, err1 := base64.StdEncoding.DecodeString(env.Salt)
	nonce, err2 := base64.StdEncoding.DecodeString(env.Nonce)
	data, err3 := base64.StdEncoding.DecodeString(env.Data)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, fmt.Errorf("the encrypted credentials block is malformed")
	}
	key, err := pbkdf2.Key(sha256.New, passphrase, salt, env.Iter, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("wrong passphrase (or the backup is corrupt)")
	}
	return plain, nil
}
