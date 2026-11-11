package core

import (
	"fmt"
	"html/template"
)

// PageContext is the data passed to templates by page-producing renderers.
type PageContext struct {
	Content  template.HTML
	URL      string
	Metadata map[string]any
}

func (p PageContext) Title() string       { return MetaString(p.Metadata, "title") }
func (p PageContext) Date() string        { return MetaString(p.Metadata, "date") }
func (p PageContext) Description() string { return MetaString(p.Metadata, "description") }

func (p PageContext) Tags() []string {
	raw, ok := p.Metadata["tags"]
	if !ok {
		return nil
	}
	switch vs := raw.(type) {
	case []any:
		tags := make([]string, 0, len(vs))
		for _, t := range vs {
			tags = append(tags, fmt.Sprint(t))
		}
		return tags
	case string:
		// `tags: go` is one tag, not a silently ignored scalar.
		if vs != "" {
			return []string{vs}
		}
	}
	return nil
}

// CollectionContext is the data passed to collection/listing templates.
type CollectionContext struct {
	Pages       []PageContext
	CurrentPage int
	TotalPages  int
	PrevURL     string
	NextURL     string
	URL         string
	Metadata    map[string]any
}

func (c CollectionContext) Title() string       { return MetaString(c.Metadata, "title") }
func (c CollectionContext) Description() string { return MetaString(c.Metadata, "description") }

// TagPageContext is the data passed to per-tag listing templates.
type TagPageContext struct {
	Tag      string
	Pages    []PageContext
	URL      string
	Metadata map[string]any
}

func (t TagPageContext) Title() string { return t.Tag }

// TagIndexEntry represents a single tag in the tag index.
type TagIndexEntry struct {
	Tag   string
	Count int
	URL   string
}

// TagIndexContext is the data passed to the tag index template.
type TagIndexContext struct {
	Tags     []TagIndexEntry
	URL      string
	Metadata map[string]any
}

func (t TagIndexContext) Title() string { return MetaString(t.Metadata, "title") }

// MetaString returns a metadata value as a string, or empty if not found.
func MetaString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprint(v)
	}
	return ""
}
