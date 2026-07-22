// Package core is the application shell: web UI, settings, history,
// export. It never executes plugins itself — everything plugin-related
// goes through the engine, so the two can evolve independently.
package core

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"orven/contract"
	"orven/internal/engine"
)

//go:embed templates/* static/*
var assets embed.FS

type Server struct {
	Engine *engine.Engine
	tmpl   *template.Template
}

func NewServer(e *engine.Engine) (*Server, error) {
	funcs := template.FuncMap{
		"day":      func(t time.Time) string { return t.Format("Monday, January 2, 2006") },
		"clock":    func(t time.Time) string { return t.Format("3:04 PM") },
		"short":    func(t time.Time) string { return t.Format("Jan 2, 3:04 PM") },
		"ago":       humanAgo,
		"staleWhen": staleWhen,
		"dur":       humanDur,
		"durc":      compactDur,
		"sizeKB": func(n int64) string {
			switch {
			case n >= 1<<20:
				return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
			default:
				return fmt.Sprintf("%d KB", (n+1023)/1024)
			}
		},
		"title":    sectionTitleStatus,
		"contains": func(xs []string, x string) bool { for _, v := range xs { if v == x { return true } }; return false },
	}
	t, err := template.New("").Funcs(funcs).ParseFS(assets, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{Engine: e, tmpl: t}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(assets))
	mux.HandleFunc("GET /{$}", s.front)
	mux.HandleFunc("GET /brief/{id}", s.brief)
	mux.HandleFunc("GET /print", s.printPreview)
	mux.HandleFunc("GET /brief/{id}/print", s.printPreview)
	mux.HandleFunc("GET /brief/{id}/export.md", s.exportMarkdown)
	mux.HandleFunc("GET /brief/{id}/export.json", s.exportJSON)
	mux.HandleFunc("GET /history", s.history)
	mux.HandleFunc("GET /plugins", s.plugins)
	mux.HandleFunc("GET /plugins/discover", s.discover)
	mux.HandleFunc("POST /plugins/discover/refresh", s.discoverRefresh)
	mux.HandleFunc("GET /plugins/install", s.installConfirm)
	mux.HandleFunc("POST /plugins/install", s.installDo)
	mux.HandleFunc("GET /plugins/{id}/uninstall", s.uninstallConfirm)
	mux.HandleFunc("POST /plugins/{id}/uninstall", s.uninstallDo)
	mux.HandleFunc("POST /settings/restore-demo", s.restoreDemo)
	mux.HandleFunc("GET /backups", s.backupsPage)
	mux.HandleFunc("POST /backups/settings", s.backupsSettingsSave)
	mux.HandleFunc("POST /backups/now", s.backupNow)
	mux.HandleFunc("POST /backups/download", s.backupDownload)
	mux.HandleFunc("GET /backups/file/{name}", s.backupFetch)
	mux.HandleFunc("POST /backups/delete", s.backupDelete)
	mux.HandleFunc("POST /backups/upload", s.backupUpload)
	mux.HandleFunc("GET /backups/restore", s.restoreConfirm)
	mux.HandleFunc("POST /backups/restore", s.restoreDo)
	mux.HandleFunc("GET /backups/restored", s.restoredPage)
	mux.HandleFunc("GET /plugins/{id}", s.pluginPage)
	mux.HandleFunc("POST /plugins/{id}/save", s.pluginSave)
	mux.HandleFunc("POST /plugins/{id}/toggle", s.pluginToggle)
	mux.HandleFunc("POST /plugins/{id}/run", s.pluginRun)
	mux.HandleFunc("GET /settings", s.settings)
	mux.HandleFunc("POST /settings/save", s.settingsSave)
	mux.HandleFunc("POST /settings/repos", s.reposEdit)
	mux.HandleFunc("POST /generate", s.generate)
	mux.HandleFunc("GET /logs", s.logs)
	// liveness for container orchestrators; no data, no auth surface
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	})
	return mux
}

func (s *Server) render(w http.ResponseWriter, page string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Page"] = page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, page+".html", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// ---- pages ----

func (s *Server) front(w http.ResponseWriter, r *http.Request) {
	brief, ok := s.Engine.Store.LatestBrief()
	data := briefView(brief)
	data["Have"], data["Now"] = ok, time.Now()
	s.render(w, "front", data)
}

func (s *Server) brief(w http.ResponseWriter, r *http.Request) {
	b, err := s.Engine.Store.Brief(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := briefView(b)
	data["Have"], data["Now"], data["Archived"] = true, time.Now(), true
	s.render(w, "front", data)
}

type coverageFailure struct{ Name, Why string }

// briefView groups a brief's sections into the three concepts the page
// presents: observed changes (the stories), the sources today's
// briefing draws on (its scope), and the sources it could not draw on.
func briefView(b engine.Brief) map[string]any {
	var stories []engine.BriefSection
	var contributed, partial, failedNames []string
	var failures []coverageFailure
	for _, sec := range b.Sections {
		if len(sec.Items) > 0 {
			stories = append(stories, sec)
		}
		switch sec.Status {
		case contract.StatusOK, contract.StatusNothing:
			contributed = append(contributed, sec.PluginName)
		case contract.StatusPartial:
			partial = append(partial, sec.PluginName)
		default:
			why := sec.Summary
			if why == "" {
				why = sectionTitleStatus(sec.Status)
			}
			failures = append(failures, coverageFailure{sec.PluginName, why})
			failedNames = append(failedNames, sec.PluginName)
		}
	}

	// The confidence line names what kept coverage from being complete;
	// the reasons live in the Coverage section below the briefing.
	var phrases []string
	if len(failedNames) > 0 {
		phrases = append(phrases, andJoin(failedNames)+" could not be checked")
	}
	if len(partial) > 0 {
		phrases = append(phrases, andJoin(partial)+" reported only partial information")
	}

	return map[string]any{
		"Brief":       b,
		"Stories":     stories,
		"Contributed": andJoin(contributed),
		"Partial":     andJoin(partial),
		"Failures":    failures,
		"Unverified":  strings.Join(phrases, "; "),
	}
}

func andJoin(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	case 2:
		return xs[0] + " and " + xs[1]
	default:
		return strings.Join(xs[:len(xs)-1], ", ") + ", and " + xs[len(xs)-1]
	}
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	type dayGroup struct {
		Day    string
		Briefs []engine.Brief
	}
	var groups []dayGroup
	for _, b := range s.Engine.Store.Briefs() {
		day := b.Generated.Format("Monday, January 2, 2006")
		if len(groups) == 0 || groups[len(groups)-1].Day != day {
			groups = append(groups, dayGroup{Day: day})
		}
		groups[len(groups)-1].Briefs = append(groups[len(groups)-1].Briefs, b)
	}
	s.render(w, "history", map[string]any{"Groups": groups, "Retention": s.Engine.Store.Settings().RetentionDays})
}

type pluginRow struct {
	P        *engine.Plugin
	Health   string
	Enabled  bool
	Interval time.Duration
	LastRun  *engine.RunRecord
	LastOK   *engine.RunRecord
	Source   string    // curated | community | bundled | manual
	Next     time.Time // next expected collection; zero = not scheduled
}

func (s *Server) pluginRows() []pluginRow {
	var rows []pluginRow
	for _, p := range s.Engine.Plugins() {
		cfg := s.Engine.Store.PluginConfig(p.Manifest.ID)
		attempt, ok := s.Engine.Store.LastRun(p.Manifest.ID)
		row := pluginRow{
			P: p, Health: s.Engine.Health(p), Enabled: cfg.Enabled,
			Interval: p.Interval(cfg), LastRun: attempt, LastOK: ok,
			Source: sourceLabel(s.Engine.Store.InstallRecord(p.Manifest.ID)),
		}
		if cfg.Enabled && p.LoadError == "" && !p.Incompatible && attempt != nil {
			row.Next = attempt.Started.Add(row.Interval)
		}
		rows = append(rows, row)
	}
	return rows
}

func (s *Server) plugins(w http.ResponseWriter, r *http.Request) {
	s.render(w, "plugins", map[string]any{
		"Rows":    s.pluginRows(),
		"Missing": s.Engine.MissingInstalls(),
		"Msg":     r.URL.Query().Get("msg"),
	})
}

func (s *Server) pluginPage(w http.ResponseWriter, r *http.Request) {
	p := s.Engine.Plugin(r.PathValue("id"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	cfg := s.Engine.Store.PluginConfig(p.Manifest.ID)
	secrets := s.Engine.Store.Secrets(p.Manifest.ID)
	secretSet := map[string]bool{}
	for k := range secrets {
		secretSet[k] = true
	}
	attempt, ok := s.Engine.Store.LastRun(p.Manifest.ID)
	runs := s.Engine.Store.Runs(p.Manifest.ID)
	if len(runs) > 10 {
		runs = runs[len(runs)-10:]
	}
	for i, j := 0, len(runs)-1; i < j; i, j = i+1, j-1 {
		runs[i], runs[j] = runs[j], runs[i]
	}
	rec := s.Engine.Store.InstallRecord(p.Manifest.ID)
	s.render(w, "plugin", map[string]any{
		"P": p, "Cfg": cfg, "SecretSet": secretSet,
		"Health": s.Engine.Health(p), "Interval": p.Interval(cfg),
		"LastRun": attempt, "LastOK": ok, "Runs": runs,
		"Rec": rec, "Source": sourceLabel(rec),
		"Msg": r.URL.Query().Get("msg"),
	})
}

// ---- actions ----

func (s *Server) pluginSave(w http.ResponseWriter, r *http.Request) {
	p := s.Engine.Plugin(r.PathValue("id"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	cfg := s.Engine.Store.PluginConfig(p.Manifest.ID)
	secrets := s.Engine.Store.Secrets(p.Manifest.ID)
	for _, f := range p.Manifest.Config {
		v := strings.TrimSpace(r.FormValue("cfg_" + f.Key))
		switch f.Type {
		case "secret":
			// write-only: empty input means "keep existing"
			if v != "" {
				secrets[f.Key] = v
			}
		case "boolean":
			cfg.Values[f.Key] = r.FormValue("cfg_"+f.Key) == "on"
		default:
			// an empty field means "use the declared default"
			if v == "" {
				delete(cfg.Values, f.Key)
			} else {
				cfg.Values[f.Key] = v
			}
		}
	}
	if iv := strings.TrimSpace(r.FormValue("interval")); iv != "" {
		if _, err := time.ParseDuration(iv); err != nil {
			http.Redirect(w, r, "/plugins/"+p.Manifest.ID+"?msg=Interval+must+look+like+30m+or+2h", http.StatusSeeOther)
			return
		}
		cfg.Interval = iv
	} else {
		cfg.Interval = ""
	}
	s.Engine.Store.SavePluginConfig(p.Manifest.ID, cfg)
	s.Engine.Store.SaveSecrets(p.Manifest.ID, secrets)
	http.Redirect(w, r, "/plugins/"+p.Manifest.ID+"?msg=Saved", http.StatusSeeOther)
}

func (s *Server) pluginToggle(w http.ResponseWriter, r *http.Request) {
	p := s.Engine.Plugin(r.PathValue("id"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	cfg := s.Engine.Store.PluginConfig(p.Manifest.ID)
	cfg.Enabled = !cfg.Enabled
	s.Engine.Store.SavePluginConfig(p.Manifest.ID, cfg)
	http.Redirect(w, r, "/plugins/"+p.Manifest.ID, http.StatusSeeOther)
}

func (s *Server) pluginRun(w http.ResponseWriter, r *http.Request) {
	p := s.Engine.Plugin(r.PathValue("id"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	msg := "Run+completed"
	if b, err := s.Engine.TryRun(p, true); err != nil {
		msg = "Run+failed:+" + strings.ReplaceAll(err.Error(), " ", "+")
	} else {
		msg = "Run+finished:+" + b.Status
	}
	http.Redirect(w, r, "/plugins/"+p.Manifest.ID+"?msg="+msg, http.StatusSeeOther)
}

func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	cfg := s.Engine.Store.Settings()
	type repoRow struct {
		URL     string
		Default bool
	}
	var repos []repoRow
	for _, u := range cfg.Repos {
		repos = append(repos, repoRow{URL: u, Default: u == engine.DefaultCatalog})
	}
	s.render(w, "settings", map[string]any{
		"S":              cfg,
		"Repos":          repos,
		"Schedule":       s.pluginRows(),
		"DemoRestorable": s.Engine.SeedAvailable("demo-activity"),
		"Msg":            r.URL.Query().Get("msg"),
		"Days":           []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	})
}

func (s *Server) settingsSave(w http.ResponseWriter, r *http.Request) {
	cfg := s.Engine.Store.Settings()
	if t := r.FormValue("brief_time"); t != "" {
		if _, err := time.Parse("15:04", t); err == nil {
			cfg.BriefTime = t
		}
	}
	cfg.BriefDays = r.Form["brief_days"]
	if len(cfg.BriefDays) == 7 {
		cfg.BriefDays = nil
	}
	if n := r.FormValue("retention"); n != "" {
		if v, err := parsePositive(n); err == nil {
			cfg.RetentionDays = v
		}
	}
	s.Engine.Store.SaveSettings(cfg)
	http.Redirect(w, r, "/settings?msg=Saved", http.StatusSeeOther)
}

func (s *Server) reposEdit(w http.ResponseWriter, r *http.Request) {
	cfg := s.Engine.Store.Settings()
	if add := strings.TrimSpace(r.FormValue("add")); add != "" {
		cfg.Repos = append(cfg.Repos, add)
	}
	if del := r.FormValue("delete"); del != "" {
		var kept []string
		for _, repo := range cfg.Repos {
			if repo != del {
				kept = append(kept, repo)
			}
		}
		cfg.Repos = kept
	}
	s.Engine.Store.SaveSettings(cfg)
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (s *Server) generate(w http.ResponseWriter, r *http.Request) {
	if _, err := s.Engine.GenerateBrief(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) logs(w http.ResponseWriter, r *http.Request) {
	s.render(w, "logs", map[string]any{"Lines": s.Engine.LogLines()})
}

// ---- helpers ----

func parsePositive(s string) (int, error) {
	var v int
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil || v <= 0 {
		return 0, fmt.Errorf("not a positive number")
	}
	return v, nil
}

// humanDur renders a duration in reader language: "30 min", "2 hr",
// "1 hr 30 min" — never Go's "30m0s" notation.
func humanDur(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%d sec", int(d.Seconds()))
	}
	if d < time.Hour {
		m, s := int(d.Minutes()), int(d.Seconds())%60
		if s == 0 {
			return fmt.Sprintf("%d min", m)
		}
		return fmt.Sprintf("%d min %d sec", m, s)
	}
	h, m := int(d.Hours()), int(d.Minutes())%60
	if m == 0 {
		return fmt.Sprintf("%d hr", h)
	}
	return fmt.Sprintf("%d hr %d min", h, m)
}

// compactDur renders a duration in the compact form users type into
// interval fields: "30m", "2h", "1h30m".
func compactDur(d time.Duration) string {
	s := d.Round(time.Second).String()
	s = strings.TrimSuffix(s, "0s")
	s = strings.TrimSuffix(s, "0m")
	if s == "" {
		return "0s"
	}
	return s
}

func humanAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "moments ago"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hr ago", int(d.Hours()))
	default:
		return t.Format("Jan 2")
	}
}

// staleWhen phrases a stale section's collection time relative to the
// briefing it appears in: "9:40 PM", "9:40 PM yesterday", or
// "July 18 at 9:40 PM".
func staleWhen(collected, generated time.Time) string {
	clock := collected.Format("3:04 PM")
	cy, cm, cd := collected.Date()
	py, pm, pd := generated.AddDate(0, 0, -1).Date()
	gy, gm, gd := generated.Date()
	switch {
	case cy == gy && cm == gm && cd == gd:
		return clock
	case cy == py && cm == pm && cd == pd:
		return clock + " yesterday"
	default:
		return collected.Format("January 2") + " at " + clock
	}
}

// sectionTitleStatus turns a section status into calm reader language.
func sectionTitleStatus(status string) string {
	switch status {
	case contract.StatusOK:
		return ""
	case contract.StatusNothing:
		return "Nothing new"
	case contract.StatusPartial:
		return "Partial information"
	case contract.StatusUnavailable:
		return "Source unavailable"
	case contract.StatusAuthFailed:
		return "Sign-in problem"
	case "no_data":
		return "No information collected"
	default:
		return "Check did not complete"
	}
}
