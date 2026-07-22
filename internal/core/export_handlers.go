package core

import (
	"encoding/json"
	"net/http"

	"orven/internal/engine"
)

// Print preview and export routes. The PDF path stays the browser's
// print-to-PDF on purpose: reliable, universal, no rendering engine
// dependency — the preview route exists so what prints is deliberate
// rather than whatever the screen happened to look like.

func (s *Server) briefFor(w http.ResponseWriter, r *http.Request) (engine.Brief, bool) {
	id := r.PathValue("id")
	if id == "" || id == "latest" {
		b, ok := s.Engine.Store.LatestBrief()
		if !ok {
			http.NotFound(w, r)
			return engine.Brief{}, false
		}
		return b, true
	}
	b, err := s.Engine.Store.Brief(id)
	if err != nil {
		http.NotFound(w, r)
		return engine.Brief{}, false
	}
	return b, true
}

func (s *Server) printPreview(w http.ResponseWriter, r *http.Request) {
	b, ok := s.briefFor(w, r)
	if !ok {
		return
	}
	data := briefView(b)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "printbrief.html", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (s *Server) exportMarkdown(w http.ResponseWriter, r *http.Request) {
	b, ok := s.briefFor(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="orven-brief-`+b.ID+`.md"`)
	w.Write([]byte(BriefMarkdown(b)))
}

func (s *Server) exportJSON(w http.ResponseWriter, r *http.Request) {
	b, ok := s.briefFor(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="orven-brief-`+b.ID+`.json"`)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(b)
}
