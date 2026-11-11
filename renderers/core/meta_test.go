package core

import (
	"reflect"
	"testing"
)

func pagesWith(field string, kv ...string) []PageContext {
	var pages []PageContext
	for i := 0; i < len(kv); i += 2 {
		meta := map[string]any{}
		if kv[i+1] != "" {
			meta[field] = kv[i+1]
		}
		pages = append(pages, PageContext{URL: kv[i], Metadata: meta})
	}
	return pages
}

func urlsOf(pages []PageContext) []string {
	var out []string
	for _, p := range pages {
		out = append(out, p.URL)
	}
	return out
}

func TestSortPages(t *testing.T) {
	tests := []struct {
		name  string
		order string
		kv    []string // url, value pairs; "" means the field is missing
		want  []string
	}{
		{
			name:  "ties break on URL ascending",
			order: "desc",
			kv:    []string{"/c", "x", "/a", "x", "/b", "y"},
			want:  []string{"/b", "/a", "/c"},
		},
		{
			name:  "dates sort by instant across UTC offsets",
			order: "asc",
			// 08:30Z then 09:00Z; as strings the order would flip.
			kv:   []string{"/later", "2024-11-03T01:00:00-08:00", "/earlier", "2024-11-03T01:30:00-07:00"},
			want: []string{"/earlier", "/later"},
		},
		{
			name:  "date-only and datetime values compare as instants",
			order: "asc",
			kv:    []string{"/b", "2024-01-14T20:00:00-08:00", "/a", "2024-01-15"},
			want:  []string{"/a", "/b"},
		},
		{
			name:  "any non-date value falls back to string order",
			order: "asc",
			kv:    []string{"/z", "zzz", "/x", "2024-11-03T01:30:00-07:00", "/y", "2024-11-03T01:00:00-08:00"},
			want:  []string{"/y", "/x", "/z"},
		},
		{
			name:  "missing values sort lowest ascending",
			order: "asc",
			kv:    []string{"/dated", "2024-01-01", "/missing", ""},
			want:  []string{"/missing", "/dated"},
		},
		{
			name:  "missing values sort last descending",
			order: "desc",
			kv:    []string{"/missing", "", "/dated", "2024-01-01"},
			want:  []string{"/dated", "/missing"},
		},
		{
			name:  "padded order keys",
			order: "asc",
			kv:    []string{"/ten", "10", "/two", "02", "/one", "01"},
			want:  []string{"/one", "/two", "/ten"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pages := pagesWith("k", tc.kv...)
			SortPages(pages, "k", tc.order)
			if got := urlsOf(pages); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadMetaFiles_Empty(t *testing.T) {
	pages, err := LoadMetaFiles("")
	if err != nil || len(pages) != 0 {
		t.Errorf("LoadMetaFiles(\"\") = %v, %v; want no pages and no error", pages, err)
	}
}

func TestTags(t *testing.T) {
	tests := []struct {
		raw  any
		want []string
	}{
		{[]any{"go", 2}, []string{"go", "2"}},
		{"go", []string{"go"}},
		{"", nil},
		{nil, nil},
	}
	for _, tc := range tests {
		p := PageContext{Metadata: map[string]any{}}
		if tc.raw != nil {
			p.Metadata["tags"] = tc.raw
		}
		if got := p.Tags(); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Tags() with %#v = %#v, want %#v", tc.raw, got, tc.want)
		}
	}
}
