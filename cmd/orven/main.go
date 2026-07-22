package main

import (
	"context"
	"encoding/json"
	"fmt"
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

// indexCLI implements `orven index <catalog-root> [catalog-name]`:
// writes the catalog's index.json to stdout.
func indexCLI(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: orven index <catalog-root> [catalog-name]")
		return 2
	}
	name := "Orven Plugins"
	if len(args) > 1 {
		name = args[1]
	}
	idx, err := engine.BuildCatalogIndex(args[0], name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "orven index:", err)
		return 1
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(idx); err != nil {
		fmt.Fprintln(os.Stderr, "orven index:", err)
		return 1
	}
	return 0
}

func main() {
	// Subcommands: `orven validate <plugin-dir>...` (docs/VALIDATOR.md)
	// and `orven index <catalog-root>` (emits a catalog's index.json).
	// Everything else serves the app.
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		os.Exit(validate.CLI(os.Args[2:]))
	}
	if len(os.Args) > 1 && os.Args[1] == "index" {
		os.Exit(indexCLI(os.Args[2:]))
	}

	dataDir := envOr("ORVEN_DATA", "data")
	pluginsDir := envOr("ORVEN_PLUGINS", "plugins")
	addr := envOr("ORVEN_ADDR", ":8420")

	store, err := engine.NewStore(dataDir)
	if err != nil {
		log.Fatalf("orven: data directory: %v", err)
	}
	eng := engine.New(store, pluginsDir)
	eng.SeedDir = os.Getenv("ORVEN_SEED")
	eng.SeedOnce()
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
