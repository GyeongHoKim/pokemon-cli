package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// pinnedSpritesSHA is the PokeAPI/sprites commit this generator pulls sprite
// images from. Re-verify and bump periodically:
// GET https://api.github.com/repos/PokeAPI/sprites/commits/master
const pinnedSpritesSHA = "bf4c47ac82c33b330e33d98b8882d1cedb2f53e7"

const (
	speciesListURL = "https://pokeapi.co/api/v2/pokemon-species?limit=1025"
	// animatedURLFmt is Pokémon Showdown's animated battle sprite: real
	// multi-frame animation, traded off against lower art fidelity than the
	// official static battle sprite. Not every species has one (~1004/1025);
	// species without it are excluded from generation entirely, not
	// fallen back to a static image.
	animatedURLFmt = "https://raw.githubusercontent.com/PokeAPI/sprites/" + pinnedSpritesSHA + "/sprites/pokemon/other/showdown/%d.gif"

	cacheDir     = ".cache/sprite-src"
	fetchWorkers = 8
	fetchRetries = 2
	httpTimeoutS = 20
)

type dexEntry struct {
	ID   int
	Name string
}

type speciesListResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func fetchDexList(client *http.Client) ([]dexEntry, error) {
	resp, err := client.Get(speciesListURL)
	if err != nil {
		return nil, fmt.Errorf("fetch species list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch species list: unexpected status %s", resp.Status)
	}

	var parsed speciesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode species list: %w", err)
	}

	entries := make([]dexEntry, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		id, err := idFromSpeciesURL(r.URL)
		if err != nil {
			return nil, fmt.Errorf("parse id for %q: %w", r.Name, err)
		}
		entries = append(entries, dexEntry{ID: id, Name: r.Name})
	}
	return entries, nil
}

// idFromSpeciesURL extracts the trailing numeric id from a PokeAPI resource
// URL such as "https://pokeapi.co/api/v2/pokemon-species/6/".
func idFromSpeciesURL(u string) (int, error) {
	trimmed := strings.TrimSuffix(u, "/")
	idx := strings.LastIndex(trimmed, "/")
	if idx < 0 {
		return 0, fmt.Errorf("no path segment in %q", u)
	}
	return strconv.Atoi(trimmed[idx+1:])
}

// fetchedSprite is the raw downloaded animated-GIF bytes for one species.
// GIF is nil when the species has no Showdown animated sprite, in which
// case the caller excludes it from generation entirely.
type fetchedSprite struct {
	ID  int
	GIF []byte // nil if unavailable
}

// fetchAll downloads (or reads from disk cache) the animated sprite for
// every entry, using a bounded worker pool.
func fetchAll(client *http.Client, entries []dexEntry) ([]fetchedSprite, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	results := make([]fetchedSprite, len(entries))
	errs := make([]error, len(entries))

	sem := make(chan struct{}, fetchWorkers)
	var wg sync.WaitGroup
	for i, e := range entries {
		wg.Add(1)
		go func(i int, e dexEntry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			gifData, err := fetchCachedOptional(client, fmt.Sprintf(animatedURLFmt, e.ID), filepath.Join(cacheDir, fmt.Sprintf("%d-animated.gif", e.ID)))
			if err != nil {
				errs[i] = fmt.Errorf("species %d (%s) animated sprite: %w", e.ID, e.Name, err)
				return
			}

			results[i] = fetchedSprite{ID: e.ID, GIF: gifData}
		}(i, e)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// fetchCachedOptional fetches url, caching bytes at cachePath, and treats
// HTTP 404 as "not available" (nil, nil) rather than an error, since not
// every species has an animated sprite.
func fetchCachedOptional(client *http.Client, url, cachePath string) ([]byte, error) {
	const missingMarker = ".missing"
	if _, err := os.Stat(cachePath + missingMarker); err == nil {
		return nil, nil
	}
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	var lastErr error
	for attempt := 0; attempt <= fetchRetries; attempt++ {
		data, status, err := httpGet(client, url)
		if err != nil {
			lastErr = err
			continue
		}
		if status == http.StatusNotFound {
			return nil, os.WriteFile(cachePath+missingMarker, nil, 0o644)
		}
		if status != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status %d for %s", status, url)
			continue
		}
		if err := os.WriteFile(cachePath, data, 0o644); err != nil {
			return nil, fmt.Errorf("write cache %s: %w", cachePath, err)
		}
		return data, nil
	}
	return nil, lastErr
}

func httpGet(client *http.Client, url string) ([]byte, int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}
