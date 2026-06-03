package steps

import (
	"io/fs"
	"testing"

	"github.com/spf13/afero"
)

// TestWalkAndCopy_Idempotency covers the 4 scenarios from the spec:
//  1. identical content → skip (no write, not in skipped)
//  2. modified content, force=false → skip (in skipped slice)
//  3. modified content, force=true → overwrite (not in skipped)
//  4. missing destination → create (normal write)

func collectFiles(fsys afero.Fs, root string) ([]string, error) {
	var files []string
	err := afero.Walk(fsys, root, func(path string, info fs.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func TestWalkAndCopy_IdempotentOnIdenticalContent(t *testing.T) {
	memFS := afero.NewMemMapFs()
	dstDir := "/dst"

	// First inject to populate files.
	injector := NewAssetInjectorWithOpts(memFS, false)
	if err := injector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("initial inject failed: %v", err)
	}

	files, _ := collectFiles(memFS, dstDir)
	if len(files) == 0 {
		t.Skip("no files in skills embed, skipping idempotency test")
	}

	// Second inject with force=false — since content is identical, skipped should be empty.
	concreteInjector, ok := injector.(*assetInjector)
	if !ok {
		t.Fatal("expected *assetInjector")
	}
	concreteInjector.skipped = nil

	if err := injector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("second inject failed: %v", err)
	}

	if len(concreteInjector.skipped) != 0 {
		t.Errorf("expected skipped=[] on identical content re-inject, got %v", concreteInjector.skipped)
	}
}

func TestWalkAndCopy_SkipsModifiedFileWhenNoForce(t *testing.T) {
	memFS := afero.NewMemMapFs()
	dstDir := "/dst"

	// First inject to populate files.
	injector := NewAssetInjectorWithOpts(memFS, false)
	if err := injector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("initial inject failed: %v", err)
	}

	files, _ := collectFiles(memFS, dstDir)
	if len(files) == 0 {
		t.Skip("no files in skills embed, skipping test")
	}
	targetFile := files[0]

	// Modify the file.
	if err := afero.WriteFile(memFS, targetFile, []byte("user-modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	concreteInjector, ok := injector.(*assetInjector)
	if !ok {
		t.Fatal("expected *assetInjector")
	}
	concreteInjector.skipped = nil

	// Re-inject with force=false — modified file should be in skipped.
	if err := injector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("re-inject failed: %v", err)
	}

	found := false
	for _, s := range concreteInjector.skipped {
		if s == targetFile {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %q in skipped, got %v", targetFile, concreteInjector.skipped)
	}

	// File content must NOT be overwritten.
	data, err := afero.ReadFile(memFS, targetFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "user-modified content" {
		t.Errorf("file was overwritten but force=false; content = %q", data)
	}
}

func TestWalkAndCopy_OverwritesModifiedFileWhenForce(t *testing.T) {
	memFS := afero.NewMemMapFs()
	dstDir := "/dst"

	// First inject to populate files.
	injector := NewAssetInjectorWithOpts(memFS, false)
	if err := injector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("initial inject failed: %v", err)
	}

	files, _ := collectFiles(memFS, dstDir)
	if len(files) == 0 {
		t.Skip("no files in skills embed, skipping test")
	}
	targetFile := files[0]

	originalData, _ := afero.ReadFile(memFS, targetFile)

	// Modify the file.
	if err := afero.WriteFile(memFS, targetFile, []byte("user-modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Re-inject with force=true using the same memFS.
	forceInjector := NewAssetInjectorWithOpts(memFS, true)
	if err := forceInjector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("force re-inject failed: %v", err)
	}

	concreteInjector, ok := forceInjector.(*assetInjector)
	if !ok {
		t.Fatal("expected *assetInjector")
	}

	// File should NOT be in skipped.
	for _, s := range concreteInjector.skipped {
		if s == targetFile {
			t.Errorf("force=true but %q is in skipped", targetFile)
		}
	}

	// File content must be restored to embedded version.
	data, err := afero.ReadFile(memFS, targetFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != string(originalData) {
		t.Errorf("expected embedded content restored, got %q", data)
	}
}

func TestWalkAndCopy_CreatesMissingFile(t *testing.T) {
	memFS := afero.NewMemMapFs()
	dstDir := "/dst/new-target"

	// Inject into a fresh directory — all files must be created.
	injector := NewAssetInjectorWithOpts(memFS, false)
	if err := injector.InjectSharedSkills(dstDir); err != nil {
		t.Fatalf("inject to new dir failed: %v", err)
	}

	concreteInjector, ok := injector.(*assetInjector)
	if !ok {
		t.Fatal("expected *assetInjector")
	}

	// skipped must be empty — new files are always written.
	if len(concreteInjector.skipped) != 0 {
		t.Errorf("new files should not be in skipped, got %v", concreteInjector.skipped)
	}

	exists, err := afero.DirExists(memFS, dstDir)
	if err != nil {
		t.Fatalf("checking dir: %v", err)
	}
	if !exists {
		t.Error("target directory was not created")
	}
}
