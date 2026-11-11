package core

import (
	"strings"

	"github.com/popovicu/pageforge/internal/templatefuncs"
)

// AbsURL joins a site base URL and a url_path in canonical form (see
// templatefuncs.CanonicalURL). Feed links and sitemap locs use it so they
// match what canonical <link> tags emit, not the file inside the tar.
func AbsURL(base, urlPath string) string {
	return strings.TrimRight(base, "/") + templatefuncs.CanonicalURL(urlPath)
}

var xmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&apos;",
)

// XMLEscape escapes the five XML predefined entities.
func XMLEscape(s string) string {
	return xmlEscaper.Replace(s)
}
