package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orven/internal/engine"
)

// Backups (Daily Care): download now, automatic schedule + retention,
// browse existing, restore — with restore as the one deliberately
// heavyweight flow, because it's the one that overwrites the present.

const uploadStagingName = "staging-restore.zip"

func (s *Server) backupsPage(w http.ResponseWriter, r *http.Request) {
	cfg := s.Engine.Store.Settings().Backup
	s.render(w, "backups", map[string]any{
		"B":             cfg,
		"Dir":           s.Engine.BackupDir(),
		"HasPassphrase": s.Engine.Store.BackupPassphrase() != "",
		"List":          s.Engine.ListBackups(),
		"Delete":        r.URL.Query().Get("delete"),
		"Msg":           r.URL.Query().Get("msg"),
	})
}

func (s *Server) backupsSettingsSave(w http.ResponseWriter, r *http.Request) {
	cfg := s.Engine.Store.Settings()
	cfg.Backup.Enabled = r.FormValue("enabled") == "on"
	if t := r.FormValue("time"); t != "" {
		if _, err := time.Parse("15:04", t); err == nil {
			cfg.Backup.Time = t
		}
	}
	cfg.Backup.Dir = strings.TrimSpace(r.FormValue("dir"))
	if v, err := parsePositive(r.FormValue("retention")); err == nil {
		cfg.Backup.Retention = v
	}
	cfg.Backup.IncludeSecrets = r.FormValue("include_secrets") == "on"
	// passphrase is write-only: empty means keep the existing one
	if p := r.FormValue("passphrase"); strings.TrimSpace(p) != "" {
		s.Engine.Store.SaveBackupPassphrase(strings.TrimSpace(p))
	}
	if cfg.Backup.IncludeSecrets && s.Engine.Store.BackupPassphrase() == "" {
		cfg.Backup.IncludeSecrets = false
		s.Engine.Store.SaveSettings(cfg)
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape("Saved — but including credentials needs a passphrase, so that stays off until you set one"), http.StatusSeeOther)
		return
	}
	s.Engine.Store.SaveSettings(cfg)
	http.Redirect(w, r, "/backups?msg=Saved", http.StatusSeeOther)
}

func (s *Server) backupNow(w http.ResponseWriter, r *http.Request) {
	name, err := s.Engine.CreateBackupFile()
	if err != nil {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	s.Engine.PruneBackups(s.Engine.Store.Settings().Backup.Retention)
	http.Redirect(w, r, "/backups?msg="+url.QueryEscape("Backup written: "+name), http.StatusSeeOther)
}

// backupDownload streams a fresh backup straight to the browser.
func (s *Server) backupDownload(w http.ResponseWriter, r *http.Request) {
	includeSecrets := r.FormValue("include_secrets") == "on"
	passphrase := r.FormValue("passphrase")
	if includeSecrets && strings.TrimSpace(passphrase) == "" {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape("Including credentials requires a passphrase"), http.StatusSeeOther)
		return
	}
	name := "orven-backup-" + time.Now().Format("20060102-150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	if err := s.Engine.WriteBackup(w, includeSecrets, passphrase); err != nil {
		s.Engine.Logf("backup download failed: %v", err)
	}
}

func (s *Server) backupFileName(raw string) (string, error) {
	name := filepath.Base(raw)
	if name != raw || !strings.HasSuffix(name, ".zip") {
		return "", fmt.Errorf("no such backup")
	}
	full := filepath.Join(s.Engine.BackupDir(), name)
	if _, err := os.Stat(full); err != nil {
		return "", fmt.Errorf("no such backup")
	}
	return full, nil
}

func (s *Server) backupFetch(w http.ResponseWriter, r *http.Request) {
	full, err := s.backupFileName(r.PathValue("name"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(full)+`"`)
	http.ServeFile(w, r, full)
}

func (s *Server) backupDelete(w http.ResponseWriter, r *http.Request) {
	full, err := s.backupFileName(r.FormValue("name"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	os.Remove(full)
	http.Redirect(w, r, "/backups?msg="+url.QueryEscape(filepath.Base(full)+" deleted"), http.StatusSeeOther)
}

// backupUpload stages an uploaded archive for the restore confirmation.
func (s *Server) backupUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256<<20)
	f, _, err := r.FormFile("archive")
	if err != nil {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape("No file was uploaded"), http.StatusSeeOther)
		return
	}
	defer f.Close()
	staging := filepath.Join(s.Engine.Store.Root, uploadStagingName)
	out, err := os.Create(staging)
	if err != nil {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	_, err = io.Copy(out, f)
	out.Close()
	if err != nil {
		os.Remove(staging)
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/backups/restore?src=upload", http.StatusSeeOther)
}

func (s *Server) restoreSourcePath(src string) (string, string, error) {
	if src == "upload" {
		full := filepath.Join(s.Engine.Store.Root, uploadStagingName)
		if _, err := os.Stat(full); err != nil {
			return "", "", fmt.Errorf("no uploaded archive is staged")
		}
		return full, "the uploaded archive", nil
	}
	full, err := s.backupFileName(src)
	if err != nil {
		return "", "", err
	}
	return full, filepath.Base(full), nil
}

func (s *Server) restoreConfirm(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("src")
	full, label, err := s.restoreSourcePath(src)
	if err != nil {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	manifest, err := engine.InspectBackup(full)
	if err != nil {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	s.render(w, "restore", map[string]any{
		"Src": src, "Label": label, "M": manifest,
		"Msg": r.URL.Query().Get("msg"),
	})
}

func (s *Server) restoreDo(w http.ResponseWriter, r *http.Request) {
	src := r.FormValue("src")
	full, _, err := s.restoreSourcePath(src)
	if err != nil {
		http.Redirect(w, r, "/backups?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if r.FormValue("ack") != "on" {
		http.Redirect(w, r, "/backups/restore?src="+url.QueryEscape(src)+"&msg="+url.QueryEscape("Confirm the overwrite to restore"), http.StatusSeeOther)
		return
	}
	manifest, merr := engine.InspectBackup(full)
	if err := s.Engine.RestoreBackup(full, r.FormValue("passphrase")); err != nil {
		http.Redirect(w, r, "/backups/restore?src="+url.QueryEscape(src)+"&msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if src == "upload" {
		os.Remove(full)
	}
	q := url.Values{}
	if merr == nil {
		q.Set("created", manifest.Created.Format(time.RFC3339))
		q.Set("briefs", fmt.Sprint(manifest.Briefs))
		q.Set("plugins", fmt.Sprint(manifest.Plugins))
		if manifest.SecretsIncluded {
			q.Set("secrets", "1")
		}
	}
	http.Redirect(w, r, "/backups/restored?"+q.Encode(), http.StatusSeeOther)
}

// restoredPage answers the three questions a restore leaves open:
// what was restored, what still needs attention, and when the user is
// fully back to the backed-up state.
func (s *Server) restoredPage(w http.ResponseWriter, r *http.Request) {
	created, _ := time.Parse(time.RFC3339, r.URL.Query().Get("created"))
	briefs, _ := parsePositive(r.URL.Query().Get("briefs"))
	plugins, _ := parsePositive(r.URL.Query().Get("plugins"))
	cfg := s.Engine.Store.Settings().Backup
	s.render(w, "restored", map[string]any{
		"Created":  created,
		"Briefs":   briefs,
		"Plugins":  plugins,
		"Secrets":  r.URL.Query().Get("secrets") == "1",
		"Missing":  s.Engine.MissingInstalls(),
		"PassWarn": cfg.IncludeSecrets && s.Engine.Store.BackupPassphrase() == "",
	})
}
