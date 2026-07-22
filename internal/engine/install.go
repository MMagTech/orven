package engine

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// InstallRecord is a plugin's provenance, written when Orven installs
// it. Identity is (catalog, plugin ID); version is release metadata
// (CONSTRAINTS.md §18-21). Managed marks that Orven placed the files
// and may therefore remove them on uninstall; a folder the user placed
// by hand has no record and its files are never deleted without the
// user's explicit acknowledgment.
type InstallRecord struct {
	PluginID  string    `json:"plugin_id"`
	Catalog   string    `json:"catalog"` // repository URL, or "bundled" for seeded content
	Status    string    `json:"status"`  // curated | community | bundled
	Publisher string    `json:"publisher"`
	Version   string    `json:"version"`
	Installed time.Time `json:"installed"`
	Managed   bool      `json:"managed"`
}

func (s *Store) InstallRecord(id string) *InstallRecord {
	var r InstallRecord
	if err := s.readJSON(filepath.Join(s.Root, "installed", id+".json"), &r); err != nil {
		return nil
	}
	return &r
}

func (s *Store) SaveInstallRecord(r InstallRecord) error {
	if err := os.MkdirAll(filepath.Join(s.Root, "installed"), 0o755); err != nil {
		return err
	}
	return s.writeJSON(filepath.Join(s.Root, "installed", r.PluginID+".json"), r)
}

func (s *Store) DeleteInstallRecord(id string) {
	os.Remove(filepath.Join(s.Root, "installed", id+".json"))
}

// Extraction limits: a plugin is a small folder; anything bigger is
// either a mistake or an attack.
const (
	maxArchiveBytes = 32 << 20
	maxPluginFiles  = 200
	maxFileBytes    = 4 << 20
)

// StageInstall downloads the repository archive and extracts the
// entry's folder into a staging directory. Nothing becomes installed
// until CommitInstall; the caller validates the staged plugin between
// the two (and DiscardStaged on failure).
func (e *Engine) StageInstall(repoURL string, entry CatalogPlugin) (string, error) {
	if p := e.Plugin(entry.ID); p != nil {
		src := "added manually"
		if rec := e.Store.InstallRecord(entry.ID); rec != nil {
			src = "installed from " + rec.Catalog
		}
		return "", fmt.Errorf("plugin id %q is already installed (%s) — an installation may contain only one plugin with a given id", entry.ID, src)
	}
	_, archiveURL, err := catalogEndpoints(repoURL)
	if err != nil {
		return "", err
	}
	resp, err := catalogHTTP.Get(archiveURL)
	if err != nil {
		return "", fmt.Errorf("the repository could not be reached")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the repository archive could not be downloaded (HTTP %d)", resp.StatusCode)
	}

	staging := filepath.Join(e.Store.Root, "staging", entry.ID)
	os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return "", err
	}
	if err := extractPluginDir(io.LimitReader(resp.Body, maxArchiveBytes), entry.Path, staging); err != nil {
		os.RemoveAll(staging)
		return "", err
	}
	if _, err := os.Stat(filepath.Join(staging, "plugin.yaml")); err != nil {
		os.RemoveAll(staging)
		return "", fmt.Errorf("the repository archive does not contain %s", entry.Path)
	}
	return staging, nil
}

// extractPluginDir extracts only wantPath's contents from a .tar.gz
// whose entries carry one leading directory component (as GitHub
// archives do). Paths are cleaned and confined to dest; symlinks and
// other special entries are ignored.
func extractPluginDir(r io.Reader, wantPath, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("the repository archive is not a gzip tarball")
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	prefix := strings.TrimSuffix(wantPath, "/") + "/"
	files := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("the repository archive is corrupt")
		}
		name := path.Clean(strings.ReplaceAll(hdr.Name, "\\", "/"))
		parts := strings.SplitN(name, "/", 2)
		if len(parts) < 2 {
			continue // the leading archive directory itself
		}
		rel := parts[1]
		if !strings.HasPrefix(rel, prefix) {
			continue
		}
		sub := path.Clean(strings.TrimPrefix(rel, prefix))
		if sub == "." || sub == "" {
			continue
		}
		if sub == ".." || strings.HasPrefix(sub, "../") || path.IsAbs(sub) {
			return fmt.Errorf("the repository archive contains an unsafe path")
		}
		target := filepath.Join(dest, filepath.FromSlash(sub))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			files++
			if files > maxPluginFiles {
				return fmt.Errorf("the plugin archive contains too many files")
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, io.LimitReader(tr, maxFileBytes))
			f.Close()
			if err != nil {
				return err
			}
		default:
			// symlinks, devices, etc. are never extracted
		}
	}
	return nil
}

// CommitInstall moves a staged (and validated) plugin into the plugins
// directory and records its provenance.
func (e *Engine) CommitInstall(staged, repoURL string, entry CatalogPlugin) error {
	dest := filepath.Join(e.PluginsDir, entry.ID)
	if _, err := os.Stat(dest); err == nil {
		os.RemoveAll(staged)
		return fmt.Errorf("plugin folder %q already exists", entry.ID)
	}
	if err := os.MkdirAll(e.PluginsDir, 0o755); err != nil {
		return err
	}
	if err := os.Rename(staged, dest); err != nil {
		// cross-device fallback
		if err2 := copyDir(staged, dest); err2 != nil {
			return err2
		}
		os.RemoveAll(staged)
	}
	rec := InstallRecord{
		PluginID: entry.ID, Catalog: repoURL, Status: entry.Status,
		Publisher: entry.Publisher, Version: entry.Version,
		Installed: time.Now(), Managed: true,
	}
	if err := e.Store.SaveInstallRecord(rec); err != nil {
		return err
	}
	e.Reload()
	e.Logf("installed plugin %s v%s from %s (%s)", entry.ID, entry.Version, repoURL, entry.Status)
	return nil
}

// DiscardStaged removes a staging directory after a failed install.
func (e *Engine) DiscardStaged(staged string) {
	if strings.Contains(staged, string(filepath.Separator)+"staging"+string(filepath.Separator)) {
		os.RemoveAll(staged)
	}
}

// Uninstall removes a plugin going forward; it never rewrites the
// past. Always removed: configuration, credentials, staged raw
// observations, and the install record. Preserved unless deleteRuns:
// the plugin's run history. The folder is deleted only when
// deleteFiles — callers must obtain the user's explicit acknowledgment
// first for folders Orven did not install. Historical briefings are
// self-contained documents and are untouched by any of this.
func (e *Engine) Uninstall(id string, deleteFiles, deleteRuns bool) error {
	p := e.Plugin(id)
	if p == nil {
		return fmt.Errorf("no such plugin")
	}
	if deleteFiles {
		dir := filepath.Clean(p.Dir)
		rel, err := filepath.Rel(e.PluginsDir, dir)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("refusing to delete files outside the plugins directory")
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	os.Remove(filepath.Join(e.Store.Root, "config", id+".json"))
	os.Remove(filepath.Join(e.Store.Root, "secrets", id+".json"))
	os.RemoveAll(filepath.Join(e.Store.Root, "observations", id))
	if deleteRuns {
		os.Remove(filepath.Join(e.Store.Root, "runs", id+".json"))
	}
	e.Store.DeleteInstallRecord(id)
	e.Reload()
	e.Logf("uninstalled plugin %s (files deleted: %v, run history deleted: %v)", id, deleteFiles, deleteRuns)
	return nil
}

// MissingInstalls returns install records whose plugin folder is not
// present — exactly the state a restore onto a fresh machine produces.
// The UI lists these so the post-restore to-do is visible in the app
// instead of remembered by the user.
func (e *Engine) MissingInstalls() []InstallRecord {
	var out []InstallRecord
	entries, _ := os.ReadDir(filepath.Join(e.Store.Root, "installed"))
	for _, ent := range entries {
		if !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(ent.Name(), ".json")
		if e.Plugin(id) != nil {
			continue
		}
		if rec := e.Store.InstallRecord(id); rec != nil {
			out = append(out, *rec)
		}
	}
	return out
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
