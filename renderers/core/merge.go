package core

// MergeMetadata deep-merges two metadata maps. The base map is the lower
// precedence (typically site-wide config); the overlay map is higher
// precedence (typically per-page front matter or YAML data). On key
// collisions overlay wins. When both sides hold a map[string]any at the
// same key, the merge recurses one level. Lists and scalars are replaced
// wholesale, not concatenated.
//
// The result is a fresh map; neither input is mutated.
func MergeMetadata(base, overlay map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(overlay))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overlay {
		if existing, ok := out[k]; ok {
			if a, aOK := existing.(map[string]any); aOK {
				if b, bOK := v.(map[string]any); bOK {
					out[k] = MergeMetadata(a, b)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
