package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:base
var FS embed.FS

// Read returns the content of an embedded file.
func Read(path string) (string, error) {
	data, err := FS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MustRead returns the content of an embedded file or panics.
func MustRead(path string) string {
	data, err := FS.ReadFile(path)
	if err != nil {
		panic("templates: " + err.Error())
	}
	return string(data)
}

// WriteEmbeddedFile reads a file from FS and writes it to the destination path.
func WriteEmbeddedFile(srcPath, destPath string) error {
	data, err := FS.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(destPath, data, 0644)
}

// CopyEmbeddedDir recursively copies a directory from FS to the physical filesystem.
func CopyEmbeddedDir(srcDir, destDir string) error {
	return fs.WalkDir(FS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return WriteEmbeddedFile(path, targetPath)
	})
}

// EmbeddedFile holds a relative path and its raw bytes from the embedded FS.
type EmbeddedFile struct {
	RelPath string
	Content []byte
}

// WalkFiles enumerates all regular files under srcDir in the embedded FS and
// returns a slice of EmbeddedFile with paths relative to srcDir.
func WalkFiles(srcDir string) ([]EmbeddedFile, error) {
	var files []EmbeddedFile
	err := fs.WalkDir(FS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		data, err := FS.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, EmbeddedFile{RelPath: rel, Content: data})
		return nil
	})
	return files, err
}
