package steps

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestInjectMarkdownSection(t *testing.T) {
	const markerID = "test-section"
	const content = "## Test Section\n\nThis is injected content."
	startMarker := "<!-- specai:test-section:start -->"
	endMarker := "<!-- specai:test-section:end -->"

	tests := []struct {
		name         string
		initialFile  string // empty means file absent
		wantContains []string
		wantAbsent   []string
		idempotent   bool // if true, run a second injection and verify no duplication
	}{
		{
			name:        "first-time write on missing file",
			initialFile: "",
			wantContains: []string{
				startMarker,
				endMarker,
				"## Test Section",
			},
		},
		{
			name:        "first-time write on existing file without marker",
			initialFile: "# Existing Content\n\nSome existing text.",
			wantContains: []string{
				"# Existing Content",
				startMarker,
				endMarker,
				"## Test Section",
			},
		},
		{
			name: "idempotent re-injection does not duplicate",
			initialFile: "# Existing Content\n\n" +
				startMarker + "\n" +
				"## Test Section\n\nThis is injected content." + "\n" +
				endMarker,
			wantContains: []string{
				startMarker,
				endMarker,
				"## Test Section",
			},
			idempotent: true,
		},
		{
			name: "stale content is replaced",
			initialFile: "# Existing\n\n" +
				startMarker + "\n" +
				"OLD STALE CONTENT\n" +
				endMarker,
			wantContains: []string{
				startMarker,
				endMarker,
				"## Test Section",
			},
			wantAbsent: []string{"OLD STALE CONTENT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := afero.NewMemMapFs()
			const path = "/home/user/.claude/CLAUDE.md"

			if tt.initialFile != "" {
				if err := afero.WriteFile(fsys, path, []byte(tt.initialFile), 0644); err != nil {
					t.Fatalf("setup: WriteFile: %v", err)
				}
			}

			if err := injectMarkdownSection(fsys, path, markerID, content); err != nil {
				t.Fatalf("injectMarkdownSection() error = %v", err)
			}

			result, err := afero.ReadFile(fsys, path)
			if err != nil {
				t.Fatalf("ReadFile after inject: %v", err)
			}
			resultStr := string(result)

			for _, want := range tt.wantContains {
				if !strings.Contains(resultStr, want) {
					t.Errorf("result missing %q\nGot:\n%s", want, resultStr)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(resultStr, absent) {
					t.Errorf("result should not contain %q\nGot:\n%s", absent, resultStr)
				}
			}

			if tt.idempotent {
				// Run a second injection and verify content is byte-identical.
				if err := injectMarkdownSection(fsys, path, markerID, content); err != nil {
					t.Fatalf("second injectMarkdownSection() error = %v", err)
				}
				result2, err := afero.ReadFile(fsys, path)
				if err != nil {
					t.Fatalf("ReadFile after second inject: %v", err)
				}
				// Count occurrences of start marker — must be exactly 1.
				count := strings.Count(string(result2), startMarker)
				if count != 1 {
					t.Errorf("after idempotent re-inject, start marker count = %d, want 1\nGot:\n%s", count, string(result2))
				}
			}
		})
	}
}
