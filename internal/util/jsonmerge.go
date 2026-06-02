package util

import "encoding/json"

// MergeJSON deep-merges src into the JSON object in dst.
// If dst is empty or missing, returns src marshaled as indented JSON.
// If dst is corrupted (invalid JSON), starts fresh with src only.
func MergeJSON(dst []byte, src map[string]any) ([]byte, error) {
	base := make(map[string]any)
	if len(dst) > 0 {
		if err := json.Unmarshal(dst, &base); err != nil {
			// Corrupted file — start fresh with src only.
			base = make(map[string]any)
		}
	}
	DeepMerge(base, src)
	return json.MarshalIndent(base, "", "  ")
}

// DeepMerge recursively merges src into dst (both map[string]any).
// Map values are merged recursively; all other values in src overwrite dst.
func DeepMerge(dst, src map[string]any) {
	for k, sv := range src {
		dv, ok := dst[k]
		if !ok {
			dst[k] = sv
			continue
		}
		dstMap, dstIsMap := dv.(map[string]any)
		srcMap, srcIsMap := sv.(map[string]any)
		if dstIsMap && srcIsMap {
			DeepMerge(dstMap, srcMap)
		} else {
			dst[k] = sv
		}
	}
}
