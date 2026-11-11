package main

import (
	"archive/tar"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func writeTar(t *testing.T, dir, name string, entries ...string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(f)
	for _, e := range entries {
		typ := byte(tar.TypeReg)
		if strings.HasSuffix(e, "/") {
			typ = tar.TypeDir
		}
		if err := tw.WriteHeader(&tar.Header{Name: e, Typeflag: typ, Mode: 0o644}); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}

// check runs sitecheck over single-file entries and label=>tar pairs.
func check(t *testing.T, files map[string]string, tars ...[2]string) error {
	t.Helper()
	dir := t.TempDir()
	m := map[string]any{"files": files, "tars": []map[string]string{}}
	for _, tr := range tars {
		m["tars"] = append(m["tars"].([]map[string]string), map[string]string{"label": tr[0], "path": tr[1]})
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return run([]string{
		"--manifest", testutil.Write(t, dir, "manifest.json", string(raw)),
		"--output", filepath.Join(dir, "stamp"),
	})
}

func TestCollisions(t *testing.T) {
	dir := t.TempDir()
	posts := writeTar(t, dir, "posts.tar", "posts/", "posts/index.html", "posts/page/2/index.html")
	postsAgain := writeTar(t, dir, "again.tar", "./posts/index.html")
	root := writeTar(t, dir, "root.tar", "robots.txt")

	tests := []struct {
		name  string
		files map[string]string
		tars  [][2]string
		want  string // substring of the error; "" means no error
	}{
		{
			name:  "disjoint inputs pass",
			files: map[string]string{"posts/a/index.html": "//:a"},
			tars:  [][2]string{{"//:posts", posts}},
		},
		{
			name:  "page collides with a collection file",
			files: map[string]string{"posts/index.html": "//:clash"},
			tars:  [][2]string{{"//:posts", posts}},
			want:  "/posts/index.html is produced by both //:clash and //:posts",
		},
		{
			name: "two collections collide despite ./ prefix",
			tars: [][2]string{{"//:posts", posts}, {"//:again", postsAgain}},
			want: "/posts/index.html is produced by both //:posts and //:again",
		},
		{
			name:  "leading slashes are normalized",
			files: map[string]string{"/robots.txt": "//:robots"},
			tars:  [][2]string{{"//:root", root}},
			want:  "/robots.txt is produced by both //:robots and //:root",
		},
		{
			name: "directories are not files",
			tars: [][2]string{{"//:posts", posts}, {"//:dirs", writeTar(t, dir, "dirs.tar", "posts/")}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := check(t, tc.files, tc.tars...)
			switch {
			case tc.want == "" && err != nil:
				t.Errorf("unexpected error: %v", err)
			case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
				t.Errorf("err = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}
