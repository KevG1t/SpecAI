package filemerge

import (
	"strings"
)

const (
	markerPrefix = "<!-- specai:"
	markerSuffix = " -->"
	closePrefix  = "<!-- /specai:"
)

var legacyPersonaFingerprints = []string{
	"## Personality",
	"Senior Architect",
	"## Rules",
}

// StripLegacyPersonaBlock removes a free-text persona block that was written
// to a markdown file outside of <!-- specai: --> markers.
func StripLegacyPersonaBlock(content string) string {
	for _, fp := range legacyPersonaFingerprints {
		if !strings.Contains(content, fp) {
			return content
		}
	}

	firstMarkerIdx := strings.Index(content, markerPrefix)

	zone := content
	if firstMarkerIdx >= 0 {
		zone = content[:firstMarkerIdx]
	}

	for _, fp := range legacyPersonaFingerprints {
		if !strings.Contains(zone, fp) {
			return content
		}
	}

	if firstMarkerIdx < 0 {
		return ""
	}

	remainder := content[firstMarkerIdx:]
	remainder = strings.TrimLeft(remainder, "\r\n")
	return remainder
}

const (
	atlBeginMarker = "<!-- BEGIN:agent-teams-lite -->"
	atlEndMarker   = "<!-- END:agent-teams-lite -->"
)

func findLineStart(s, needle string) int {
	offset := 0
	for {
		idx := strings.Index(s[offset:], needle)
		if idx < 0 {
			return -1
		}
		absIdx := offset + idx
		if absIdx == 0 || s[absIdx-1] == '\n' {
			return absIdx
		}
		offset = absIdx + 1
		if offset >= len(s) {
			return -1
		}
	}
}

func removeLineStartMarkers(content, marker string) string {
	for {
		idx := findLineStart(content, marker)
		if idx < 0 {
			return content
		}
		end := idx + len(marker)
		if end < len(content) && content[end] == '\r' {
			end++
		}
		if end < len(content) && content[end] == '\n' {
			end++
		}
		content = content[:idx] + content[end:]
	}
}

// StripLegacyATLBlock removes the legacy Agent Teams Lite block wrapped in
// <!-- BEGIN:agent-teams-lite --> / <!-- END:agent-teams-lite --> markers.
func StripLegacyATLBlock(content string) string {
	for {
		beginIdx := findLineStart(content, atlBeginMarker)
		if beginIdx < 0 {
			break
		}

		searchFrom := beginIdx + len(atlBeginMarker)
		relEndIdx := findLineStart(content[searchFrom:], atlEndMarker)
		if relEndIdx < 0 {
			break
		}
		endIdx := searchFrom + relEndIdx

		before := content[:beginIdx]
		after := content[endIdx+len(atlEndMarker):]

		before = strings.TrimRight(before, "\r\n")
		after = strings.TrimLeft(after, "\r\n")

		if before == "" && after == "" {
			content = ""
			continue
		}

		var sb strings.Builder
		if before != "" {
			sb.WriteString(before)
			sb.WriteString("\n")
		}
		if after != "" {
			if before != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(after)
		}

		content = sb.String()
	}

	content = removeLineStartMarkers(content, atlEndMarker)
	content = removeLineStartMarkers(content, atlBeginMarker)

	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}
	return content
}

func openMarker(sectionID string) string {
	return markerPrefix + sectionID + markerSuffix
}

func closeMarker(sectionID string) string {
	return closePrefix + sectionID + markerSuffix
}

func stripOrphanMarkers(content, open, close string) string {
	for {
		openIdx := strings.Index(content, open)
		closeIdx := strings.Index(content, close)

		switch {
		case openIdx < 0 && closeIdx < 0:
			return content

		case openIdx < 0 && closeIdx >= 0:
			content = content[:closeIdx] + content[closeIdx+len(close):]

		case openIdx >= 0 && closeIdx < 0:
			content = content[:openIdx] + content[openIdx+len(open):]

		case closeIdx < openIdx:
			content = content[:closeIdx] + content[closeIdx+len(close):]

		default:
			return content
		}
	}
}

// InjectMarkdownSection replaces or appends a marked section in a markdown file.
// Markers use HTML comments: <!-- specai:SECTION_ID --> ... <!-- /specai:SECTION_ID -->
// If the section already exists, its content is replaced.
// If it doesn't exist, it's appended at the end.
// Content outside markers is never touched.
// If content is empty, the section (including markers) is removed.
func InjectMarkdownSection(existing, sectionID, content string) string {
	open := openMarker(sectionID)
	close := closeMarker(sectionID)

	existing = stripOrphanMarkers(existing, open, close)

	openIdx := strings.Index(existing, open)
	closeIdx := strings.Index(existing, close)

	if openIdx >= 0 && closeIdx >= 0 && closeIdx > openIdx {
		if content == "" {
			before := existing[:openIdx]
			after := existing[closeIdx+len(close):]

			if len(after) > 0 && after[0] == '\n' {
				after = after[1:]
			}
			result := strings.TrimRight(before, "\n")
			if after != "" {
				if result != "" {
					result += "\n"
				}
				result += after
			} else if result != "" {
				result += "\n"
			}
			return result
		}

		before := existing[:openIdx]
		after := existing[closeIdx+len(close):]

		var sb strings.Builder
		sb.WriteString(before)
		sb.WriteString(open)
		sb.WriteString("\n")
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString(close)
		sb.WriteString(after)
		return sb.String()
	}

	if content == "" {
		return existing
	}

	var sb strings.Builder
	sb.WriteString(existing)
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		sb.WriteString("\n")
	}
	if existing != "" {
		sb.WriteString("\n")
	}
	sb.WriteString(open)
	sb.WriteString("\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(close)
	sb.WriteString("\n")
	return sb.String()
}
