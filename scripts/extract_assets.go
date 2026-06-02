package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ExtractAssets copies files from src to dst.
func ExtractAssets(src string, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if filepath.Ext(path) == ".go" {
			return nil // Skip .go files
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		if err := os.WriteFile(dstPath, content, info.Mode()); err != nil {
			return fmt.Errorf("failed to write file %s: %w", dstPath, err)
		}

		return nil
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <src> [dst]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  dst defaults to internal/assets/\n")
		os.Exit(1)
	}

	src := os.Args[1]
	dst := "internal/assets"
	if len(os.Args) >= 3 {
		dst = os.Args[2]
	}

	if err := ExtractAssets(src, dst); err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting assets: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Assets extracted successfully.")
}
