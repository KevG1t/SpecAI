package filemerge

import (
	"fmt"
	"strings"
)

// UpsertCodexSDDMemoryBlock removes any existing [mcp_servers.sdd-memory] block from
// the given TOML content and appends a fresh block with the canonical sdd-memory
// MCP entry (including --tools=agent). All other sections are preserved.
//
// sddMemoryCmd is the command string to use (e.g. an absolute path).
// If sddMemoryCmd is empty, it falls back to "sdd-memory".
func UpsertCodexSDDMemoryBlock(content, sddMemoryCmd string) string {
	if sddMemoryCmd == "" {
		sddMemoryCmd = "sdd-memory"
	}
	// Escape backslashes for TOML double-quoted strings (Windows paths).
	escapedCmd := strings.ReplaceAll(sddMemoryCmd, `\`, `\\`)
	block := "[mcp_servers.sdd-memory]\ncommand = \"" + escapedCmd + "\"\nargs = [\"mcp\", \"--tools=agent\"]"
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	var kept []string
	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "[mcp_servers.sdd-memory]" {
			i++
			for i < len(lines) {
				next := strings.TrimSpace(lines[i])
				if strings.HasPrefix(next, "[") && strings.HasSuffix(next, "]") {
					break
				}
				i++
			}
			continue
		}

		kept = append(kept, lines[i])
		i++
	}

	base := strings.TrimSpace(strings.Join(kept, "\n"))
	if base == "" {
		return block + "\n"
	}

	return base + "\n\n" + block + "\n"
}

// UpsertTopLevelTOMLString inserts or replaces a top-level key = "value" pair
// in TOML content. The key is placed before the first [section] header so it
// remains a top-level (non-table) setting. Existing occurrences of the key are
// removed before inserting the new value (idempotent).
func UpsertTopLevelTOMLString(content, key, value string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	lineValue := fmt.Sprintf("%s = %q", key, value)

	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+" ") || strings.HasPrefix(trimmed, key+"=") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	insertAt := len(cleaned)
	for i, line := range cleaned {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			insertAt = i
			break
		}
	}

	var out []string
	out = append(out, cleaned[:insertAt]...)
	out = append(out, lineValue)
	out = append(out, cleaned[insertAt:]...)

	return strings.TrimSpace(strings.Join(out, "\n")) + "\n"
}
