// sitecheck validates that no two site_build inputs ship the same path. It
// runs as site_build's validation action: single-file entries (deps,
// static) are compared against every file inside the collection tars, and
// the tars against each other. pkg_tar would otherwise keep the first copy
// and only print a warning.
//
// Flags:
//
//	--manifest   JSON: {"files": {path: label}, "tars": [{"label", "path"}]}
//	--output     Stamp file written when the check passes
package main

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

type manifest struct {
	Files map[string]string `json:"files"`
	Tars  []struct {
		Label string `json:"label"`
		Path  string `json:"path"`
	} `json:"tars"`
}

func run(args []string) error {
	fs := flag.NewFlagSet("sitecheck", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "JSON manifest of single-file entries and collection tars")
	output := fs.String("output", "", "Stamp file written on success")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *output == "" {
		return fmt.Errorf("--manifest and --output are required")
	}

	raw, err := os.ReadFile(*manifestPath)
	if err != nil {
		return fmt.Errorf("reading manifest: %w", err)
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("parsing manifest: %w", err)
	}

	owners := map[string]string{}
	for p, label := range m.Files {
		owners[normalize(p)] = label
	}
	var conflicts []string
	for _, t := range m.Tars {
		names, err := tarFiles(t.Path)
		if err != nil {
			return err
		}
		for _, n := range names {
			n = normalize(n)
			if prev, ok := owners[n]; ok && prev != t.Label {
				conflicts = append(conflicts, fmt.Sprintf("/%s is produced by both %s and %s", n, prev, t.Label))
			}
			owners[n] = t.Label
		}
	}
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return fmt.Errorf("site_build URL collisions:\n  %s", strings.Join(conflicts, "\n  "))
	}
	return os.WriteFile(*output, nil, 0o644)
}

// normalize maps the spellings pkg_tar treats as one member ("/a", "./a",
// "a") to a single key.
func normalize(p string) string {
	return strings.TrimPrefix(path.Clean("/"+p), "/")
}

// tarFiles lists the regular files in a tar archive.
func tarFiles(p string) ([]string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var names []string
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return names, nil
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		if hdr.FileInfo().Mode().IsRegular() {
			names = append(names, hdr.Name)
		}
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "sitecheck: %v\n", err)
		os.Exit(1)
	}
}
