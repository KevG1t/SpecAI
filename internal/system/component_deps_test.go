package system

import (
	"context"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
)

// allKnownComponents is the exhaustive list of ComponentIDs defined in model/types.go.
// ADD NEW ComponentIDs HERE when they are added to model/types.go.
var allKnownComponents = []model.ComponentID{
	model.ComponentSDDMemory,
	model.ComponentSDD,
	model.ComponentSkills,
	model.ComponentContext7,
	model.ComponentPersona,
	model.ComponentPermission,
	model.ComponentTheme,
	model.ComponentClaudeTheme,
	model.ComponentOpenCodeArgentinaLogo,
	model.ComponentNotion,
	model.ComponentJira,
}

// depFreeComponents lists components that have no system prerequisites.
var depFreeComponents = map[model.ComponentID]bool{
	model.ComponentSDD:                   true,
	model.ComponentSkills:                true,
	model.ComponentPersona:               true,
	model.ComponentPermission:            true,
	model.ComponentTheme:                 true,
	model.ComponentClaudeTheme:           true,
	model.ComponentOpenCodeArgentinaLogo: true,
}

// allowedDepNames is the set of valid dependency names for the componentDepMap entries.
var allowedDepNames = map[string]bool{
	"git":  true,
	"curl": true,
	"node": true,
	"npm":  true,
	"brew": true,
	"go":   true,
	"bash": true,
}

// --- T-1: Mapping table completeness ---

func TestComponentDepMapCompleteness(t *testing.T) {
	for _, id := range allKnownComponents {
		_, inDepMap := componentDepMap[id]
		_, inDepFree := depFreeComponents[id]
		if !inDepMap && !inDepFree {
			t.Errorf("ComponentID %q is not in componentDepMap and not in depFreeComponents — add it to one", id)
		}
	}
}

func TestComponentDepMapAllowedDepNames(t *testing.T) {
	for id, deps := range componentDepMap {
		for _, rd := range deps {
			if !allowedDepNames[rd.Name] {
				t.Errorf("componentDepMap[%q] has dep %q which is not in allowedDepNames", id, rd.Name)
			}
		}
	}
}

// --- T-2: Builder — empty / nil selection ---

func TestBuildComponentDepTreeNilSelection(t *testing.T) {
	tree := BuildComponentDepTree(context.Background(), nil, PlatformProfile{})
	if tree != nil {
		t.Fatalf("expected nil tree for nil selection, got %v", tree)
	}
}

func TestBuildComponentDepTreeEmptySelection(t *testing.T) {
	tree := BuildComponentDepTree(context.Background(), []model.ComponentID{}, PlatformProfile{})
	if tree != nil {
		t.Fatalf("expected nil tree for empty selection, got %v", tree)
	}
}

// --- T-3: Builder — single component, dep-free and with deps ---

func TestBuildComponentDepTreeDepFreeComponent(t *testing.T) {
	orig := detectFn
	detectFn = func(_ context.Context, dep Dependency) Dependency {
		dep.Installed = true
		dep.Version = "1.0.0"
		return dep
	}
	t.Cleanup(func() { detectFn = orig })

	tree := BuildComponentDepTree(context.Background(), []model.ComponentID{model.ComponentSDD}, PlatformProfile{})
	if tree == nil {
		t.Fatal("expected non-nil tree for dep-free component selection")
	}
	rdSlice, ok := tree[model.ComponentSDD]
	if !ok {
		t.Fatal("expected ComponentSDD key in result tree")
	}
	if len(rdSlice) != 0 {
		t.Fatalf("expected empty slice for dep-free component, got %v", rdSlice)
	}
}

func TestBuildComponentDepTreeSingleComponent(t *testing.T) {
	orig := detectFn
	detectFn = func(_ context.Context, dep Dependency) Dependency {
		dep.Installed = true
		dep.Version = "1.0.0"
		return dep
	}
	t.Cleanup(func() { detectFn = orig })

	tree := BuildComponentDepTree(context.Background(), []model.ComponentID{model.ComponentSDDMemory}, PlatformProfile{})
	if tree == nil {
		t.Fatal("expected non-nil tree")
	}
	rdSlice, ok := tree[model.ComponentSDDMemory]
	if !ok {
		t.Fatal("expected ComponentSDDMemory key in result tree")
	}
	// SDDMemory declares curl (required) and brew (optional).
	if len(rdSlice) != 2 {
		t.Fatalf("expected 2 RequiredDep entries for SDDMemory, got %d", len(rdSlice))
	}
	names := map[string]bool{}
	for _, rd := range rdSlice {
		names[rd.Name] = true
	}
	if !names["curl"] {
		t.Error("expected curl in SDDMemory deps")
	}
	if !names["brew"] {
		t.Error("expected brew in SDDMemory deps")
	}
}

// --- T-4: Builder — overlapping deps (deduplication) ---

func TestBuildComponentDepTreeDeduplication(t *testing.T) {
	callCount := 0
	orig := detectFn
	detectFn = func(_ context.Context, dep Dependency) Dependency {
		if dep.Name == "curl" {
			callCount++
		}
		dep.Installed = true
		dep.Version = "1.0.0"
		return dep
	}
	t.Cleanup(func() { detectFn = orig })

	// Both ComponentSDDMemory and a hypothetical second entry would share curl.
	// Use ComponentSDDMemory + ComponentContext7 — SDDMemory has curl, Context7 has none.
	// To test dedup with TWO components sharing curl we need another component with curl.
	// Since only SDDMemory has curl in componentDepMap, use SDDMemory twice via a temporary override.
	// Instead, inject a temporary componentDepMap entry.
	origMap := componentDepMap
	componentDepMap = map[model.ComponentID][]RequiredDep{
		model.ComponentSDDMemory: {{Name: "curl", Required: true}},
		model.ComponentContext7:  {{Name: "curl", Required: false}},
	}
	t.Cleanup(func() { componentDepMap = origMap })

	_ = BuildComponentDepTree(context.Background(), []model.ComponentID{model.ComponentSDDMemory, model.ComponentContext7}, PlatformProfile{})

	if callCount != 1 {
		t.Fatalf("expected detectFn to be called exactly once for curl (dedup), got %d", callCount)
	}
}

func TestBuildComponentDepTreeNeededByPopulated(t *testing.T) {
	orig := detectFn
	detectFn = func(_ context.Context, dep Dependency) Dependency {
		dep.Installed = true
		return dep
	}
	t.Cleanup(func() { detectFn = orig })

	origMap := componentDepMap
	componentDepMap = map[model.ComponentID][]RequiredDep{
		model.ComponentSDDMemory: {{Name: "curl", Required: true}},
		model.ComponentContext7:  {{Name: "curl", Required: false}},
	}
	t.Cleanup(func() { componentDepMap = origMap })

	tree := BuildComponentDepTree(context.Background(),
		[]model.ComponentID{model.ComponentSDDMemory, model.ComponentContext7},
		PlatformProfile{})

	if tree == nil {
		t.Fatal("expected non-nil tree")
	}

	// Both entries should have curl's NeededBy containing both components.
	sddDeps := tree[model.ComponentSDDMemory]
	if len(sddDeps) != 1 || sddDeps[0].Name != "curl" {
		t.Fatalf("unexpected SDDMemory deps: %v", sddDeps)
	}
	neededBy := sddDeps[0].NeededBy
	if len(neededBy) != 2 {
		t.Fatalf("expected 2 NeededBy entries for curl, got %d: %v", len(neededBy), neededBy)
	}
}

// --- T-5: Builder — unknown ComponentID ---

func TestBuildComponentDepTreeUnknownComponent(t *testing.T) {
	orig := detectFn
	detectFn = func(_ context.Context, dep Dependency) Dependency {
		dep.Installed = true
		return dep
	}
	t.Cleanup(func() { detectFn = orig })

	tree := BuildComponentDepTree(context.Background(),
		[]model.ComponentID{model.ComponentID("unknown-agent")},
		PlatformProfile{})

	if len(tree) != 0 {
		t.Fatalf("expected zero entries for unknown component, got %v", tree)
	}
}

// --- T-6: NeededBy reverse index ---

func TestNeededByNilTree(t *testing.T) {
	result := NeededBy(nil)
	if result == nil {
		t.Fatal("expected non-nil map for nil input")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map for nil input, got %v", result)
	}
}

func TestNeededByEmptyTree(t *testing.T) {
	result := NeededBy(ComponentDepTree{})
	if result == nil {
		t.Fatal("expected non-nil map for empty tree")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map for empty tree, got %v", result)
	}
}

func TestNeededByGroupsAndSorts(t *testing.T) {
	// Simulate: SDDMemory→[curl], Context7→[curl] so curl has 2 components.
	tree := ComponentDepTree{
		model.ComponentSDDMemory: {
			{Name: "curl", Required: true},
			{Name: "brew", Required: false},
		},
		model.ComponentContext7: {
			{Name: "curl", Required: false},
		},
	}

	result := NeededBy(tree)

	curlComponents, ok := result["curl"]
	if !ok {
		t.Fatal("expected curl in NeededBy result")
	}
	if len(curlComponents) != 2 {
		t.Fatalf("expected 2 components for curl, got %d: %v", len(curlComponents), curlComponents)
	}

	// Verify sorted order (context7 < sdd-memory lexicographically).
	if string(curlComponents[0]) >= string(curlComponents[1]) {
		t.Fatalf("NeededBy slice not sorted: %v", curlComponents)
	}

	brewComponents, ok := result["brew"]
	if !ok {
		t.Fatal("expected brew in NeededBy result")
	}
	if len(brewComponents) != 1 || brewComponents[0] != model.ComponentSDDMemory {
		t.Fatalf("expected [SDDMemory] for brew, got %v", brewComponents)
	}
}

// --- T-7: DetectionResult zero value ---

func TestDetectionResultZeroValue(t *testing.T) {
	var result DetectionResult
	if result.ComponentTree != nil {
		t.Fatalf("expected ComponentTree to be nil in zero-value DetectionResult, got %v", result.ComponentTree)
	}
}
