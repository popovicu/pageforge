package core

import (
	"reflect"
	"testing"
)

func TestMergeMetadata_OverlayWinsOnTopLevel(t *testing.T) {
	base := map[string]any{
		"title":       "site title",
		"description": "site description",
	}
	overlay := map[string]any{
		"title": "page title",
	}
	got := MergeMetadata(base, overlay)
	want := map[string]any{
		"title":       "page title",
		"description": "site description",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("merge:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMergeMetadata_NestedMapsRecurse(t *testing.T) {
	base := map[string]any{
		"site": map[string]any{
			"base_url": "https://example.com",
			"author":   "Alice",
		},
	}
	overlay := map[string]any{
		"site": map[string]any{
			"author": "Bob",
		},
		"title": "Hello",
	}
	got := MergeMetadata(base, overlay)
	want := map[string]any{
		"site": map[string]any{
			"base_url": "https://example.com",
			"author":   "Bob",
		},
		"title": "Hello",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("merge:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMergeMetadata_ListsReplaceWholesale(t *testing.T) {
	base := map[string]any{
		"tags": []any{"a", "b"},
	}
	overlay := map[string]any{
		"tags": []any{"c"},
	}
	got := MergeMetadata(base, overlay)
	want := map[string]any{
		"tags": []any{"c"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("merge:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMergeMetadata_ScalarOverridesNestedMap(t *testing.T) {
	// Type mismatch: overlay scalar replaces base map (no recursion).
	base := map[string]any{
		"site": map[string]any{"author": "Alice"},
	}
	overlay := map[string]any{
		"site": "scalar",
	}
	got := MergeMetadata(base, overlay)
	want := map[string]any{"site": "scalar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("merge:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMergeMetadata_DoesNotMutateInputs(t *testing.T) {
	base := map[string]any{"k": "base"}
	overlay := map[string]any{"k": "overlay"}
	_ = MergeMetadata(base, overlay)
	if base["k"] != "base" {
		t.Errorf("base mutated: got %#v", base)
	}
	if overlay["k"] != "overlay" {
		t.Errorf("overlay mutated: got %#v", overlay)
	}
}

func TestMergeMetadata_NilInputs(t *testing.T) {
	got := MergeMetadata(nil, map[string]any{"k": "v"})
	want := map[string]any{"k": "v"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("nil base:\n got: %#v\nwant: %#v", got, want)
	}

	got = MergeMetadata(map[string]any{"k": "v"}, nil)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("nil overlay:\n got: %#v\nwant: %#v", got, want)
	}
}
