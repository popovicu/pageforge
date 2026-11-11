package core

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SplitFrontMatter separates YAML front matter from the body of a file.
// Front matter opens with a first line of "---" and closes at the next line
// that is exactly "---"; trailing whitespace and CRLF endings are allowed,
// so "----" or "--- x" never close it. Without both delimiters the whole
// file is body and metadata is empty.
func SplitFrontMatter(data []byte) (map[string]any, []byte, error) {
	s := string(data)
	open, rest, ok := strings.Cut(s, "\n")
	if !ok || !isDelimiter(open) {
		return map[string]any{}, data, nil
	}

	var frontMatter, body string
	closed := false
	for pos := 0; pos <= len(rest); {
		line, after, more := strings.Cut(rest[pos:], "\n")
		if isDelimiter(line) {
			frontMatter, body, closed = rest[:pos], after, true
			break
		}
		if !more {
			break
		}
		pos += len(line) + 1
	}
	if !closed {
		return map[string]any{}, data, nil
	}

	var metadata map[string]any
	if err := yaml.Unmarshal([]byte(frontMatter), &metadata); err != nil {
		return nil, nil, fmt.Errorf("invalid YAML front matter: %w", err)
	}
	if metadata == nil {
		metadata = map[string]any{}
	}

	return stringifyTimes(metadata).(map[string]any), []byte(body), nil
}

func isDelimiter(line string) bool {
	return strings.TrimRight(line, " \t\r") == "---"
}

// LoadYAML reads a YAML file and returns its contents as a map.
func LoadYAML(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var data map[string]any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	if data == nil {
		data = map[string]any{}
	}

	return stringifyTimes(data).(map[string]any), nil
}

// stringifyTimes replaces the time.Time values yaml.v3 produces for unquoted
// dates with RFC 3339 strings, the same text the .meta.json sidecar holds.
// Page templates and listings then see one representation, and string
// helpers like formatTime accept it.
func stringifyTimes(v any) any {
	switch x := v.(type) {
	case time.Time:
		return x.Format(time.RFC3339Nano)
	case map[string]any:
		for k, e := range x {
			x[k] = stringifyTimes(e)
		}
	case []any:
		for i, e := range x {
			x[i] = stringifyTimes(e)
		}
	}
	return v
}
