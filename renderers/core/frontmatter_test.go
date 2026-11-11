package core

import (
	"reflect"
	"testing"
)

func TestSplitFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantMeta map[string]any
		wantBody string
	}{
		{"no front matter", "# Hi\n", map[string]any{}, "# Hi\n"},
		{"front matter", "---\ntitle: Hi\n---\nbody\n", map[string]any{"title": "Hi"}, "body\n"},
		{"unterminated block is body", "---\ntitle: Hi\nbody\n", map[string]any{}, "---\ntitle: Hi\nbody\n"},
		{"empty front matter", "---\n---\nbody\n", map[string]any{}, "body\n"},
		{"closing delimiter at EOF", "---\ntitle: Hi\n---", map[string]any{"title": "Hi"}, ""},
		{"CRLF line endings", "---\r\ntitle: Hi\r\n---\r\nbody\r\n", map[string]any{"title": "Hi"}, "body\r\n"},
		{"trailing whitespace on delimiters", "--- \ntitle: Hi\n---\t\nbody\n", map[string]any{"title": "Hi"}, "body\n"},
		{"longer rule does not close", "---\ntitle: Hi\n----\n", map[string]any{}, "---\ntitle: Hi\n----\n"},
		{"text after dashes does not open", "--- x\ntitle: Hi\n---\n", map[string]any{}, "--- x\ntitle: Hi\n---\n"},
		{"unquoted date becomes RFC 3339", "---\ndate: 2026-04-15\n---\n", map[string]any{"date": "2026-04-15T00:00:00Z"}, ""},
		{"quoted date stays as written", "---\ndate: \"2026-04-15\"\n---\n", map[string]any{"date": "2026-04-15"}, ""},
		{
			"nested dates are converted",
			"---\nevents: [2026-01-02]\nmeta: {when: 2026-01-03}\n---\n",
			map[string]any{
				"events": []any{"2026-01-02T00:00:00Z"},
				"meta":   map[string]any{"when": "2026-01-03T00:00:00Z"},
			},
			"",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta, body, err := SplitFrontMatter([]byte(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(meta, tc.wantMeta) {
				t.Errorf("meta:\n got: %#v\nwant: %#v", meta, tc.wantMeta)
			}
			if string(body) != tc.wantBody {
				t.Errorf("body: got %q, want %q", body, tc.wantBody)
			}
		})
	}
}

func TestSplitFrontMatter_InvalidYAML(t *testing.T) {
	if _, _, err := SplitFrontMatter([]byte("---\ntitle: [\n---\n")); err == nil {
		t.Error("expected an error for malformed front matter")
	}
}
