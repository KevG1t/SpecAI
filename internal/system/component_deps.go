package system

import (
	"context"
	"sort"
	"sync"

	"github.com/KevG1t/SpecAI/internal/model"
)

// RequiredDep describes a system prerequisite as declared by a component.
// It is a declaration, not a detection result — Installed/Version are not fields here.
// NeededBy is populated by BuildComponentDepTree (reverse index).
type RequiredDep struct {
	Name       string              // binary name matching Dependency.Name ("node", "git", …)
	MinVersion string              // minimum acceptable version, empty = any
	Required   bool                // true = hard requirement, false = optional/recommended
	NeededBy   []model.ComponentID // populated by BuildComponentDepTree (reverse index)
}

// ComponentDepTree maps each selected ecosystem component to the system deps it requires.
// The word "tree" is retained for PRD continuity; the structure is a grouped flat list,
// not a recursive tree. The zero value (nil map) means "no selection active".
type ComponentDepTree map[model.ComponentID][]RequiredDep

// componentDepMap maps each ComponentID to the system deps it requires.
// Dep-free components (file-copy only) have an empty slice.
// Components not in this map are treated as unknown and silently skipped by the builder.
var componentDepMap = map[model.ComponentID][]RequiredDep{
	model.ComponentSDDMemory: {
		{Name: "curl", Required: true},
		{Name: "brew", Required: false}, // optional; darwin-only at render time
	},
	// Remote MCP — no local binary required.
	model.ComponentContext7: {},
	model.ComponentNotion:   {},
	model.ComponentJira:     {},
	// File-copy only — also in depFreeComponents for the completeness test.
	model.ComponentSDD:                   {},
	model.ComponentSkills:                {},
	model.ComponentPersona:               {},
	model.ComponentPermission:            {},
	model.ComponentTheme:                 {},
	model.ComponentClaudeTheme:           {},
	model.ComponentOpenCodeArgentinaLogo: {},
}

// detectFn is injectable for tests so they don't exec real system commands.
// Production code uses detectSingleDep directly.
var detectFn = func(ctx context.Context, dep Dependency) Dependency {
	return detectSingleDep(ctx, dep)
}

// BuildComponentDepTree builds a ComponentDepTree for the provided selection.
// Returns nil when selection is empty or nil.
// Unknown ComponentIDs (not in componentDepMap) are silently skipped.
// Dep-free components appear as keys with an empty []RequiredDep{} slice.
// Detection calls run concurrently and are de-duplicated by dep name.
// NeededBy is populated on each RequiredDep in the result (reverse index, sorted).
func BuildComponentDepTree(ctx context.Context, selection []model.ComponentID, profile PlatformProfile) ComponentDepTree {
	if len(selection) == 0 {
		return nil
	}

	// Collect only the selected components that exist in the map.
	type compEntry struct {
		id   model.ComponentID
		deps []RequiredDep
	}
	var entries []compEntry
	for _, id := range selection {
		decls, ok := componentDepMap[id]
		if !ok {
			continue // unknown component — silently skip
		}
		entries = append(entries, compEntry{id: id, deps: decls})
	}

	// Build union of unique dep names across all selected components.
	depNames := map[string]bool{}
	for _, e := range entries {
		for _, rd := range e.deps {
			depNames[rd.Name] = true
		}
	}

	// Build a lookup from dep name → Dependency definition (for DetectCmd, InstallHint, etc.).
	defMap := map[string]Dependency{}
	for _, d := range defineDependencies(profile) {
		defMap[d.Name] = d
	}

	// Run detectFn concurrently for each unique dep name.
	detected := make(map[string]Dependency, len(depNames))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name := range depNames {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			def, ok := defMap[n]
			if !ok {
				// Dep not in the canonical list (e.g. brew on non-darwin) —
				// construct a minimal Dependency so detection can still run.
				def = Dependency{Name: n}
			}
			result := detectFn(ctx, def)
			mu.Lock()
			detected[n] = result
			mu.Unlock()
		}(name)
	}
	wg.Wait()

	// Assemble result tree.
	tree := make(ComponentDepTree, len(entries))
	for _, e := range entries {
		if len(e.deps) == 0 {
			tree[e.id] = []RequiredDep{} // dep-free: explicit empty slice, not nil
			continue
		}
		rdSlice := make([]RequiredDep, len(e.deps))
		for i, rd := range e.deps {
			detectedDep := detected[rd.Name]
			rdSlice[i] = RequiredDep{
				Name:       rd.Name,
				MinVersion: rd.MinVersion,
				Required:   rd.Required,
				// NeededBy populated below after reverse-index pass
			}
			_ = detectedDep // detection result available; callers use DependencyReport for Installed/Version
		}
		tree[e.id] = rdSlice
	}

	// Build reverse index: dep name → []ComponentID.
	reverseIdx := map[string][]model.ComponentID{}
	for id, rdSlice := range tree {
		for _, rd := range rdSlice {
			reverseIdx[rd.Name] = append(reverseIdx[rd.Name], id)
		}
	}
	// Sort each slice for deterministic output.
	for name := range reverseIdx {
		ids := reverseIdx[name]
		sort.Slice(ids, func(i, j int) bool {
			return string(ids[i]) < string(ids[j])
		})
		reverseIdx[name] = ids
	}

	// Write NeededBy back into each RequiredDep.
	for id, rdSlice := range tree {
		for i, rd := range rdSlice {
			rdSlice[i].NeededBy = reverseIdx[rd.Name]
		}
		tree[id] = rdSlice
	}

	return tree
}

// NeededBy returns a map from dependency name to the sorted slice of ComponentIDs
// that require it, derived from the provided tree.
// Returns an empty non-nil map when tree is nil or empty.
func NeededBy(tree ComponentDepTree) map[string][]model.ComponentID {
	result := map[string][]model.ComponentID{}
	if len(tree) == 0 {
		return result
	}

	for id, rdSlice := range tree {
		for _, rd := range rdSlice {
			result[rd.Name] = append(result[rd.Name], id)
		}
	}

	for name := range result {
		ids := result[name]
		sort.Slice(ids, func(i, j int) bool {
			return string(ids[i]) < string(ids[j])
		})
		result[name] = ids
	}

	return result
}
