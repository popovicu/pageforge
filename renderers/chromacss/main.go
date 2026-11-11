// chromacss emits a Chroma CSS stylesheet for a named style. The produced
// file matches the class names that the markdown renderer's Chroma formatter
// writes into rendered HTML, so loading this stylesheet styles fenced code
// blocks.
//
// Flags:
//
//	--style              Chroma style name (e.g. "monokai", "gruvbox")
//	--output             Output .css file path
//	--with-all-classes   Emit the full class set even when redundant (default true)
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func run(args []string) error {
	fs := flag.NewFlagSet("chromacss", flag.ContinueOnError)
	flagStyle := fs.String("style", "", "Chroma style name")
	flagOutput := fs.String("output", "", "Output CSS file path")
	flagAllClasses := fs.Bool("with-all-classes", true, "Emit the full Chroma class set")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *flagStyle == "" || *flagOutput == "" {
		return fmt.Errorf("--style and --output are required")
	}

	style := styles.Get(*flagStyle)
	if style == styles.Fallback && *flagStyle != "fallback" {
		names := strings.Join(styles.Names(), ", ")
		return fmt.Errorf("unknown chroma style %q; valid styles: %s", *flagStyle, names)
	}

	f := html.New(
		html.WithClasses(true),
		html.WithAllClasses(*flagAllClasses),
	)

	out, err := os.Create(*flagOutput)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	if err := f.WriteCSS(out, style); err != nil {
		out.Close()
		return fmt.Errorf("writing CSS: %w", err)
	}
	return out.Close()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "chromacss: %v\n", err)
		os.Exit(1)
	}
}
