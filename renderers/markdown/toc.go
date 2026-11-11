package main

import (
	"fmt"
	"html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// tocExtension turns a literal `## <Marker>` heading (default: "Table of
// contents") into the declaration of a TOC anchor: the heading itself is
// preserved, and a <details><summary>...</summary><ul>...</ul></details>
// block listing every subsequent H2/H3 is inserted right after it.
//
// Heading anchor ids come from parser.WithAutoHeadingID() — the upstream
// parser sets node.AttributeString("id") during parsing of each ATX heading,
// so by the time our AST transformer runs the ids are already in place.
type tocExtension struct {
	Marker  string
	Summary string
}

func (e tocExtension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&tocTransformer{marker: e.Marker}, 900),
	))
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&tocRenderer{summary: e.Summary}, 500),
	))
}

var tocKind = ast.NewNodeKind("PageforgeTOC")

type tocItem struct {
	id       string
	text     string
	children []tocItem
}

type tocBlock struct {
	ast.BaseBlock
	items []tocItem
}

func (n *tocBlock) Kind() ast.NodeKind { return tocKind }

func (n *tocBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

type tocTransformer struct {
	marker string
}

func (t *tocTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()

	var marker *ast.Heading
	var items []tocItem

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		text := strings.TrimSpace(extractHeadingText(h, src))
		if marker == nil {
			if h.Level == 2 && strings.EqualFold(text, t.marker) {
				marker = h
			}
			return ast.WalkContinue, nil
		}
		if h.Level != 2 && h.Level != 3 {
			return ast.WalkContinue, nil
		}
		idAttr, ok := h.AttributeString("id")
		if !ok {
			return ast.WalkContinue, nil
		}
		id, ok := idAttr.([]byte)
		if !ok {
			return ast.WalkContinue, nil
		}
		item := tocItem{id: string(id), text: text}
		switch h.Level {
		case 2:
			items = append(items, item)
		case 3:
			if len(items) == 0 {
				items = append(items, tocItem{children: []tocItem{item}})
			} else {
				last := &items[len(items)-1]
				last.children = append(last.children, item)
			}
		}
		return ast.WalkContinue, nil
	})

	if marker == nil || len(items) == 0 {
		return
	}

	toc := &tocBlock{items: items}
	parent := marker.Parent()
	parent.InsertAfter(parent, marker, toc)
}

// extractHeadingText concatenates plaintext from a heading's inline children.
// Inline formatting (code spans, emphasis) is flattened to plain text — the
// TOC favors readability over byte-for-byte preservation of the heading
// markup.
func extractHeadingText(h *ast.Heading, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(h, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Text:
			b.Write(v.Segment.Value(src))
		case *ast.String:
			b.Write(v.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

type tocRenderer struct {
	summary string
}

func (r *tocRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(tocKind, r.render)
}

func (r *tocRenderer) render(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	toc := n.(*tocBlock)
	fmt.Fprintf(w, `<details class="pageforge-toc"><summary>%s</summary>`, html.EscapeString(r.summary))
	w.WriteByte('\n')
	w.WriteString("<ul>\n")
	for _, top := range toc.items {
		w.WriteString("<li>")
		writeTocLink(w, top)
		if len(top.children) > 0 {
			w.WriteString("\n<ul>\n")
			for _, sub := range top.children {
				w.WriteString("<li>")
				writeTocLink(w, sub)
				w.WriteString("</li>\n")
			}
			w.WriteString("</ul>\n")
		}
		w.WriteString("</li>\n")
	}
	w.WriteString("</ul>\n")
	w.WriteString("</details>\n")
	return ast.WalkSkipChildren, nil
}

func writeTocLink(w util.BufWriter, it tocItem) {
	if it.id == "" {
		// H3 promoted to top level when no preceding H2 exists — render the
		// children, no link for the synthetic parent.
		fmt.Fprint(w, html.EscapeString(it.text))
		return
	}
	fmt.Fprintf(w, `<a href="#%s">%s</a>`, html.EscapeString(it.id), html.EscapeString(it.text))
}
