package core

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/popovicu/pageforge/internal/templatefuncs"
)

// LoadMetaFiles reads .meta.json sidecar files and returns PageContext entries.
// An empty list yields no pages, so aggregating zero pages is not an error.
func LoadMetaFiles(commaSeparated string) ([]PageContext, error) {
	if commaSeparated == "" {
		return nil, nil
	}
	var pages []PageContext
	for _, path := range strings.Split(commaSeparated, ",") {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		var meta map[string]any
		if err := json.Unmarshal(raw, &meta); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		pages = append(pages, PageContext{
			URL:      MetaString(meta, "url_path"),
			Metadata: meta,
		})
	}
	return pages, nil
}

// SortPages sorts pages by a metadata field in the given order. When every
// non-empty value of the field is a date (see templatefuncs.ParseDate),
// pages sort by instant, so mixed UTC offsets order correctly; otherwise
// values compare as strings. Missing values sort lowest either way. Ties
// break on URL ascending so every listing orders equal keys the same way.
func SortPages(pages []PageContext, field, order string) {
	type item struct {
		page PageContext
		key  string
		at   time.Time
	}
	items := make([]item, len(pages))
	chrono := true
	for i, p := range pages {
		it := item{page: p, key: MetaString(p.Metadata, field)}
		if it.key != "" {
			t, ok := templatefuncs.ParseDate(it.key)
			chrono = chrono && ok
			it.at = t
		}
		items[i] = it
	}
	compare := func(a, b item) int {
		if chrono && a.key != "" && b.key != "" {
			return a.at.Compare(b.at)
		}
		return strings.Compare(a.key, b.key)
	}
	slices.SortStableFunc(items, func(a, b item) int {
		c := compare(a, b)
		if c == 0 {
			return strings.Compare(a.page.URL, b.page.URL)
		}
		if order == "asc" {
			return c
		}
		return -c
	})
	for i, it := range items {
		pages[i] = it.page
	}
}

// CloseTar finalizes a tar archive and its file, reporting the first error.
// A failed footer write otherwise leaves a truncated archive behind a
// successful action.
func CloseTar(tw *tar.Writer, f *os.File) error {
	if err := tw.Close(); err != nil {
		f.Close()
		return fmt.Errorf("finalizing tar: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing tar: %w", err)
	}
	return nil
}

// ParseExtraMetadata unmarshals a JSON string into a metadata map.
func ParseExtraMetadata(jsonStr string) (map[string]any, error) {
	var extra map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &extra); err != nil {
		return nil, fmt.Errorf("parsing extra-metadata: %w", err)
	}
	return extra, nil
}

// WriteTarEntry writes a single file entry to a tar writer.
func WriteTarEntry(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("writing tar header for %s: %w", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("writing tar entry %s: %w", name, err)
	}
	return nil
}
