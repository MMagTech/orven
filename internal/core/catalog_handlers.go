package core

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"orven/internal/engine"
	"orven/internal/validate"
)

// The Discover tier: browse configured repositories, see what each
// offers with full provenance and permissions, and install. Labels
// follow CONSTRAINTS §17 — valid is never trusted, curated is an
// editorial recommendation, community and third-party say what they
// are on every surface.

type discoverEntry struct {
	engine.CatalogPlugin
	Installed bool
}

type discoverRepo struct {
	URL     string
	Name    string
	Default bool
	Fetched time.Time
	Err     string
	Plugins []discoverEntry
}

func (s *Server) discoverRepos(force bool) []discoverRepo {
	var out []discoverRepo
	for _, repoURL := range s.Engine.Store.Settings().Repos {
		row := discoverRepo{URL: repoURL, Default: repoURL == engine.DefaultCatalog, Name: repoDisplayName(repoURL, "")}
		cat, err := s.Engine.Catalog(repoURL, force)
		if err != nil {
			row.Err = err.Error()
			out = append(out, row)
			continue
		}
		row.Name = repoDisplayName(repoURL, cat.Index.Name)
		row.Fetched = cat.Fetched
		for _, p := range cat.Index.Plugins {
			row.Plugins = append(row.Plugins, discoverEntry{CatalogPlugin: p, Installed: s.Engine.Plugin(p.ID) != nil})
		}
		out = append(out, row)
	}
	return out
}

func repoDisplayName(repoURL, indexName string) string {
	if indexName != "" {
		return indexName
	}
	if u, err := url.Parse(repoURL); err == nil && u.Host != "" {
		return u.Host + u.Path
	}
	return repoURL
}

func (s *Server) discover(w http.ResponseWriter, r *http.Request) {
	s.render(w, "discover", map[string]any{
		"Repos": s.discoverRepos(false),
		"Msg":   r.URL.Query().Get("msg"),
	})
}

func (s *Server) discoverRefresh(w http.ResponseWriter, r *http.Request) {
	s.discoverRepos(true)
	http.Redirect(w, r, "/plugins/discover?msg=Repositories+refreshed", http.StatusSeeOther)
}

// findCatalogEntry locates a plugin in a configured repository's
// cached index; installs are only ever an explicit pick of a specific
// plugin from a specific repository (§20).
func (s *Server) findCatalogEntry(repoURL, id string) (engine.CatalogPlugin, string, bool) {
	for _, configured := range s.Engine.Store.Settings().Repos {
		if configured != repoURL {
			continue
		}
		cat, err := s.Engine.Catalog(repoURL, false)
		if err != nil {
			return engine.CatalogPlugin{}, "", false
		}
		for _, p := range cat.Index.Plugins {
			if p.ID == id {
				return p, repoDisplayName(repoURL, cat.Index.Name), true
			}
		}
	}
	return engine.CatalogPlugin{}, "", false
}

func (s *Server) installConfirm(w http.ResponseWriter, r *http.Request) {
	repoURL, id := r.URL.Query().Get("repo"), r.URL.Query().Get("id")
	entry, repoName, ok := s.findCatalogEntry(repoURL, id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.render(w, "install", map[string]any{
		"Entry": entry, "RepoURL": repoURL, "RepoName": repoName,
		"Default":   repoURL == engine.DefaultCatalog,
		"Installed": s.Engine.Plugin(id) != nil,
	})
}

func (s *Server) installDo(w http.ResponseWriter, r *http.Request) {
	repoURL, id := r.FormValue("repo"), r.FormValue("id")
	entry, repoName, ok := s.findCatalogEntry(repoURL, id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	staged, err := s.Engine.StageInstall(repoURL, entry)
	if err != nil {
		s.render(w, "install", map[string]any{
			"Entry": entry, "RepoURL": repoURL, "RepoName": repoName,
			"Default": repoURL == engine.DefaultCatalog, "Error": err.Error(),
		})
		return
	}
	// The validator gate applies to every install, whatever the source.
	findings := validate.Dir(staged)
	var errs []string
	for _, f := range findings {
		if f.Severity == "ERROR" {
			errs = append(errs, f.Where+": "+f.Message)
		}
	}
	if len(errs) > 0 {
		s.Engine.DiscardStaged(staged)
		s.render(w, "install", map[string]any{
			"Entry": entry, "RepoURL": repoURL, "RepoName": repoName,
			"Default": repoURL == engine.DefaultCatalog,
			"Error":   "This plugin failed validation and was not installed.",
			"Errs":    errs,
		})
		return
	}
	if err := s.Engine.CommitInstall(staged, repoURL, entry); err != nil {
		s.Engine.DiscardStaged(staged)
		s.render(w, "install", map[string]any{
			"Entry": entry, "RepoURL": repoURL, "RepoName": repoName,
			"Default": repoURL == engine.DefaultCatalog, "Error": err.Error(),
		})
		return
	}
	http.Redirect(w, r, "/plugins/"+id+"?msg=Installed+—+configure+and+enable+it+below", http.StatusSeeOther)
}

func (s *Server) uninstallConfirm(w http.ResponseWriter, r *http.Request) {
	p := s.Engine.Plugin(r.PathValue("id"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	rec := s.Engine.Store.InstallRecord(p.Manifest.ID)
	s.render(w, "uninstall", map[string]any{
		"P":       p,
		"Managed": rec != nil && rec.Managed,
		"Rec":     rec,
		"Dir":     p.Dir,
		"HasRuns": len(s.Engine.Store.Runs(p.Manifest.ID)) > 0,
		"Msg":     r.URL.Query().Get("msg"),
	})
}

func (s *Server) uninstallDo(w http.ResponseWriter, r *http.Request) {
	p := s.Engine.Plugin(r.PathValue("id"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	id := p.Manifest.ID
	rec := s.Engine.Store.InstallRecord(id)
	managed := rec != nil && rec.Managed
	if !managed && r.FormValue("ack_delete") != "on" {
		http.Redirect(w, r, "/plugins/"+id+"/uninstall?msg=Confirm+file+deletion+to+uninstall+this+manually+added+plugin",
			http.StatusSeeOther)
		return
	}
	deleteRuns := r.FormValue("delete_runs") == "on"
	if err := s.Engine.Uninstall(id, true, deleteRuns); err != nil {
		http.Redirect(w, r, "/plugins/"+id+"/uninstall?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	if id == demoPluginID {
		s.Engine.FinishOnboarding() // uninstalling the demo ends the first-run experience
	}
	http.Redirect(w, r, "/plugins?msg="+url.QueryEscape(p.Manifest.Name+" was uninstalled. Historical briefings are unchanged."), http.StatusSeeOther)
}

func (s *Server) restoreDemo(w http.ResponseWriter, r *http.Request) {
	if err := s.Engine.RestoreSeed("demo-activity"); err != nil {
		http.Redirect(w, r, "/settings?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/plugins/demo-activity?msg=Demo+restored", http.StatusSeeOther)
}

// sourceLabel is the provenance pill shown wherever a plugin appears.
func sourceLabel(rec *engine.InstallRecord) string {
	if rec == nil {
		return "manual"
	}
	switch rec.Status {
	case "bundled":
		return "bundled"
	case "curated", "community":
		return rec.Status
	}
	return strings.ToLower(strings.TrimSpace(rec.Status))
}
