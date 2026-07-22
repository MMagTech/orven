package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DefaultCatalog is the repository shipped in a fresh installation's
// settings. Being the default is a display distinction only ("default
// catalog" vs "third-party repository") — trust rules are identical.
const DefaultCatalog = "https://github.com/MMagTech/orven-plugins"

// CatalogPlugin is one entry in a repository's published index.json.
type CatalogPlugin struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Publisher   string   `json:"publisher"`
	Version     string   `json:"version"`
	Status      string   `json:"status"` // curated | community
	Path        string   `json:"path"`   // plugin folder within the repository
	Permissions []string `json:"permissions"`
}

// CatalogIndex is the machine-readable catalog a plugin repository
// publishes as index.json at its root.
type CatalogIndex struct {
	Name      string          `json:"name,omitempty"`
	Generated time.Time       `json:"generated,omitzero"`
	Plugins   []CatalogPlugin `json:"plugins"`
}

// CachedCatalog is a fetched index plus when it was fetched; Discover
// reads from this cache and refreshes on demand.
type CachedCatalog struct {
	URL     string       `json:"url"`
	Fetched time.Time    `json:"fetched"`
	Index   CatalogIndex `json:"index"`
}

// catalogEndpoints resolves a configured repository URL to its index
// and archive URLs. GitHub repositories get raw/codeload forms; any
// other base URL is expected to serve /index.json and /archive.tar.gz
// itself, which is also how self-hosted catalogs work.
func catalogEndpoints(repoURL string) (indexURL, archiveURL string, err error) {
	u := strings.TrimSuffix(strings.TrimSpace(repoURL), "/")
	if u == "" {
		return "", "", fmt.Errorf("empty repository URL")
	}
	if owner, repo, ok := githubRepo(u); ok {
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/index.json", owner, repo),
			fmt.Sprintf("https://codeload.github.com/%s/%s/tar.gz/refs/heads/main", owner, repo),
			nil
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return "", "", fmt.Errorf("repository URL must start with https:// (got %q)", repoURL)
	}
	return u + "/index.json", u + "/archive.tar.gz", nil
}

func githubRepo(u string) (owner, repo string, ok bool) {
	for _, prefix := range []string{"https://github.com/", "http://github.com/"} {
		if strings.HasPrefix(u, prefix) {
			parts := strings.Split(strings.TrimPrefix(u, prefix), "/")
			if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
				return parts[0], strings.TrimSuffix(parts[1], ".git"), true
			}
		}
	}
	return "", "", false
}

var catalogHTTP = &http.Client{Timeout: 30 * time.Second}

func (s *Store) catalogCachePath(repoURL string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(repoURL)))
	return filepath.Join(s.Root, "catalogs", hex.EncodeToString(h[:8])+".json")
}

// Catalog returns the cached index for repoURL, fetching when there is
// no cache yet or when force is set (the Refresh action).
func (e *Engine) Catalog(repoURL string, force bool) (CachedCatalog, error) {
	path := e.Store.catalogCachePath(repoURL)
	var cached CachedCatalog
	if !force {
		if err := e.Store.readJSON(path, &cached); err == nil {
			return cached, nil
		}
	}
	indexURL, _, err := catalogEndpoints(repoURL)
	if err != nil {
		return cached, err
	}
	resp, err := catalogHTTP.Get(indexURL)
	if err != nil {
		return cached, fmt.Errorf("the repository could not be reached")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cached, fmt.Errorf("the repository does not publish a plugin index (HTTP %d for index.json)", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return cached, err
	}
	var idx CatalogIndex
	if err := json.Unmarshal(body, &idx); err != nil {
		return cached, fmt.Errorf("the repository's index.json is malformed")
	}
	cached = CachedCatalog{URL: repoURL, Fetched: time.Now(), Index: idx}
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := e.Store.writeJSON(path, cached); err != nil {
		e.Logf("catalog cache write failed for %s: %v", repoURL, err)
	}
	return cached, nil
}

// BuildCatalogIndex walks a catalog repository working tree
// (plugins/curated, plugins/community) and produces its index.json.
// Used by `orven index` and the catalog repository's CI.
func BuildCatalogIndex(root, name string) (CatalogIndex, error) {
	idx := CatalogIndex{Name: name, Generated: time.Now().UTC().Truncate(time.Second)}
	for _, status := range []string{"curated", "community"} {
		dir := filepath.Join(root, "plugins", status)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, ent := range entries {
			if !ent.IsDir() {
				continue
			}
			p := LoadPlugin(filepath.Join(dir, ent.Name()))
			if p == nil {
				continue
			}
			if p.LoadError != "" {
				return idx, fmt.Errorf("%s/%s: %s", status, ent.Name(), p.LoadError)
			}
			idx.Plugins = append(idx.Plugins, CatalogPlugin{
				ID:          p.Manifest.ID,
				Name:        p.Manifest.Name,
				Description: strings.TrimSpace(p.Manifest.Description),
				Publisher:   p.Manifest.Publisher,
				Version:     p.Manifest.Version,
				Status:      status,
				Path:        "plugins/" + status + "/" + ent.Name(),
				Permissions: p.Manifest.Permissions,
			})
		}
	}
	sort.Slice(idx.Plugins, func(i, j int) bool { return idx.Plugins[i].ID < idx.Plugins[j].ID })
	return idx, nil
}
