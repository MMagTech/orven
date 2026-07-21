package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"orven/internal/core"
	"orven/internal/engine"
	"orven/internal/validate"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// Subcommand: `orven validate <plugin-dir>...` — the contributor
	// tool specified in docs/VALIDATOR.md. Everything else serves the app.
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		os.Exit(validate.CLI(os.Args[2:]))
	}

	dataDir := envOr("ORVEN_DATA", "data")
	pluginsDir := envOr("ORVEN_PLUGINS", "plugins")
	addr := envOr("ORVEN_ADDR", ":8420")

	store, err := engine.NewStore(dataDir)
	if err != nil {
		log.Fatalf("orven: data directory: %v", err)
	}
	eng := engine.New(store, pluginsDir)
	eng.StartScheduler(context.Background())

	srv, err := core.NewServer(eng)
	if err != nil {
		log.Fatalf("orven: templates: %v", err)
	}
	eng.Logf("orven listening on %s (data=%s plugins=%s)", addr, dataDir, pluginsDir)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
