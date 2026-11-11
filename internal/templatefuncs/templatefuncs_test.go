package templatefuncs

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		value  string
		want   string
	}{
		{
			name:   "valid RFC3339 reformatted with friendly layout",
			layout: "January 2, 2006",
			value:  "2026-04-23T10:30:00Z",
			want:   "April 23, 2026",
		},
		{
			name:   "valid RFC3339 with offset",
			layout: "2006-01-02",
			value:  "2026-04-23T10:30:00+02:00",
			want:   "2026-04-23",
		},
		{
			name:   "empty value returns empty unchanged",
			layout: "January 2, 2006",
			value:  "",
			want:   "",
		},
		{
			name:   "bad RFC3339 input returns value unchanged",
			layout: "January 2, 2006",
			value:  "not a date",
			want:   "not a date",
		},
		{
			name:   "date-only YYYY-MM-DD is formatted",
			layout: "January 2, 2006",
			value:  "2026-04-23",
			want:   "April 23, 2026",
		},
		{
			name:   "date-like but invalid returns value unchanged",
			layout: "January 2, 2006",
			value:  "2026-13-40",
			want:   "2026-13-40",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatTime(tc.layout, tc.value)
			if got != tc.want {
				t.Errorf("formatTime(%q, %q) = %q, want %q", tc.layout, tc.value, got, tc.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		value string
		want  string // RFC 3339 in UTC, or "" when not a date
	}{
		{"2026-04-23T10:30:00-07:00", "2026-04-23T17:30:00Z"},
		{"2026-04-23", "2026-04-23T00:00:00Z"},
		{"", ""},
		{"01", ""},
		{"April 23, 2026", ""},
	}
	for _, tc := range tests {
		got, ok := ParseDate(tc.value)
		if tc.want == "" {
			if ok {
				t.Errorf("ParseDate(%q) = %v, want not a date", tc.value, got)
			}
			continue
		}
		if !ok || got.UTC().Format(time.RFC3339) != tc.want {
			t.Errorf("ParseDate(%q) = %v, %v; want %s", tc.value, got, ok, tc.want)
		}
	}
}

func TestCanonicalURL(t *testing.T) {
	tests := map[string]string{
		"/posts/x/index.html": "/posts/x/",
		"/index.html":         "/",
		"posts/x/index.html":  "/posts/x/",
		"/about.html":         "/about.html",
		"/posts/":             "/posts/",
		"/myindex.html":       "/myindex.html",
	}
	for in, want := range tests {
		if got := CanonicalURL(in); got != want {
			t.Errorf("CanonicalURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		s      string
		want   bool
	}{
		{"matching prefix", "/series/", "/series/foo", true},
		{"non-matching prefix", "/series/", "/posts/foo", false},
		{"empty prefix always matches", "", "/anything", true},
		{"empty string with empty prefix", "", "", true},
		{"empty string with non-empty prefix", "/x", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasPrefix(tc.prefix, tc.s)
			if got != tc.want {
				t.Errorf("hasPrefix(%q, %q) = %v, want %v", tc.prefix, tc.s, got, tc.want)
			}
		})
	}
}

func TestHasSuffix(t *testing.T) {
	tests := []struct {
		name   string
		suffix string
		s      string
		want   bool
	}{
		{"matching suffix", ".html", "index.html", true},
		{"non-matching suffix", ".html", "index.md", false},
		{"empty suffix always matches", "", "/anything", true},
		{"empty string with empty suffix", "", "", true},
		{"empty string with non-empty suffix", ".html", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasSuffix(tc.suffix, tc.s)
			if got != tc.want {
				t.Errorf("hasSuffix(%q, %q) = %v, want %v", tc.suffix, tc.s, got, tc.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive + positive", 2, 3, 5},
		{"zero + positive", 0, 7, 7},
		{"negative + positive", -4, 10, 6},
		{"negative + negative", -3, -2, -5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := add(tc.a, tc.b); got != tc.want {
				t.Errorf("add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive - positive", 10, 3, 7},
		{"zero - positive", 0, 7, -7},
		{"result is negative", 3, 10, -7},
		{"negative - negative", -3, -2, -1},
		{"negative - positive", -4, 10, -14},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sub(tc.a, tc.b); got != tc.want {
				t.Errorf("sub(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestDefaultRegistersAllFunctions(t *testing.T) {
	fm := Default()
	for _, name := range []string{"formatTime", "canonicalURL", "hasPrefix", "hasSuffix", "add", "sub"} {
		if _, ok := fm[name]; !ok {
			t.Errorf("Default() is missing %q", name)
		}
	}
}
