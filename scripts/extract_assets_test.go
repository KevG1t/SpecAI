package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractAndReplace(t *testing.T) {
	// Create a temporary source directory
	srcDir, err := os.MkdirTemp("", "gentle-ai-src")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create a temporary destination directory
	dstDir, err := os.MkdirTemp("", "gentle-ai-dst")
	if err != nil {
		t.Fatalf("Failed to create temp dst dir: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Create a mock skill file with "engram" in it
	skillDir := filepath.Join(srcDir, "sdd-tasks")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	
	srcFile := filepath.Join(skillDir, "SKILL.md")
	content := "This is a test file using engram for memory. Also engram config."
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write mock skill file: %v", err)
	}

	// Create a mock .go file that should be ignored
	goFile := filepath.Join(skillDir, "ignored.go")
	if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write mock go file: %v", err)
	}

	// Call the function we're testing
	err = ExtractAndReplace(srcDir, dstDir)
	if err != nil {
		t.Fatalf("ExtractAndReplace failed: %v", err)
	}

	// Verify the destination file exists
	dstFile := filepath.Join(dstDir, "sdd-tasks", "SKILL.md")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist in destination", dstFile)
	}

	// Verify the content has "engram" replaced with "sdd-memory"
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	expectedContent := "This is a test file using sdd-memory for memory. Also sdd-memory config."
	if string(dstContent) != expectedContent {
		t.Errorf("Content replacement failed.\nExpected: %s\nGot: %s", expectedContent, string(dstContent))
	}

	// Verify the .go file was NOT copied
	dstGoFile := filepath.Join(dstDir, "sdd-tasks", "ignored.go")
	if _, err := os.Stat(dstGoFile); !os.IsNotExist(err) {
		t.Fatalf("Expected file %s to NOT exist in destination, but it does", dstGoFile)
	}
}
