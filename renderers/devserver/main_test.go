package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func TestHandlerNeverServesFromCache(t *testing.T) {
	dir := t.TempDir()
	testutil.Write(t, dir, "index.html", "fresh")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// A browser revalidating with the tarball's fixed mtime must still get
	// the current file, not a 304.
	req.Header.Set("If-Modified-Since", "Sat, 01 Jan 2050 00:00:00 GMT")
	rec := httptest.NewRecorder()
	handler(dir).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != "fresh" {
		t.Errorf("got %d %q, want 200 \"fresh\"", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}
