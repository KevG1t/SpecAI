package steps

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// injectMarkdownSection injects content into a file using HTML marker comments for idempotency.
//
// Marker format:
//
//	<!-- specai:{markerID}:start -->
//	{content}
//	<!-- specai:{markerID}:end -->
//
// Algorithm:
//   - If markers are found, the inner content is replaced (idempotent).
//   - If markers are not found, the full fenced block (markers + content) is appended.
//   - If the file does not exist, it is created.
func injectMarkdownSection(fsys afero.Fs, filePath, markerID, content string) error {
	startMarker := fmt.Sprintf("<!-- specai:%s:start -->", markerID)
	endMarker := fmt.Sprintf("<!-- specai:%s:end -->", markerID)

	// Read existing content (or start empty if file does not exist).
	existing, err := afero.ReadFile(fsys, filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	block := startMarker + "\n" + content + "\n" + endMarker

	var result []byte
	startIdx := bytes.Index(existing, []byte(startMarker))
	endIdx := bytes.Index(existing, []byte(endMarker))

	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		// Replace content between the markers (inclusive).
		endOfMarker := endIdx + len(endMarker)
		result = append(existing[:startIdx], []byte(block)...)
		result = append(result, existing[endOfMarker:]...)
	} else {
		// Append the new section.
		if len(existing) > 0 {
			result = append(existing, '\n', '\n')
		}
		result = append(result, []byte(block)...)
	}

	if err := fsys.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(filePath), err)
	}
	return writeFileAtomic(fsys, filePath, result, 0644)
}
