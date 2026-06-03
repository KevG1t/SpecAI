package mcp

import (
	"fmt"

	"github.com/KevG1t/SpecAI/internal/versions"
)

var defaultContext7ServerJSON = []byte(fmt.Sprintf("{\n  \"command\": \"npx\",\n  \"args\": [\n    \"-y\",\n    \"--package=@upstash/context7-mcp@%s\",\n    \"--\",\n    \"context7-mcp\"\n  ]\n}\n", versions.Context7MCP))

var defaultContext7OverlayJSON = []byte(fmt.Sprintf("{\n  \"mcpServers\": {\n    \"context7\": {\n      \"command\": \"npx\",\n      \"args\": [\n        \"-y\",\n        \"--package=@upstash/context7-mcp@%s\",\n        \"--\",\n        \"context7-mcp\"\n      ]\n    }\n  }\n}\n", versions.Context7MCP))

var openCodeContext7OverlayJSON = []byte("{\n  \"mcp\": {\n    \"context7\": {\n      \"__replace__\": {\n        \"type\": \"remote\",\n        \"url\": \"https://mcp.context7.com/mcp\",\n        \"enabled\": true\n      }\n    }\n  }\n}\n")

var openClawContext7OverlayJSON = []byte(fmt.Sprintf("{\n  \"mcp\": {\n    \"servers\": {\n      \"context7\": {\n        \"command\": \"npx\",\n        \"args\": [\n          \"-y\",\n          \"--package=@upstash/context7-mcp@%s\",\n          \"--\",\n          \"context7-mcp\"\n        ]\n      }\n    }\n  }\n}\n", versions.Context7MCP))

var vsCodeContext7OverlayJSON = []byte("{\n  \"servers\": {\n    \"context7\": {\n      \"type\": \"http\",\n      \"url\": \"https://mcp.context7.com/mcp\"\n    }\n  }\n}\n")

var antigravityContext7OverlayJSON = []byte("{\n  \"mcpServers\": {\n    \"context7\": {\n      \"__replace__\": {\n        \"serverUrl\": \"https://mcp.context7.com/mcp\"\n      }\n    }\n  }\n}\n")

var kimiContext7OverlayJSON = []byte("{\n  \"mcpServers\": {\n    \"context7\": {\n      \"__replace__\": {\n        \"transport\": \"http\",\n        \"url\": \"https://mcp.context7.com/mcp\"\n      }\n    }\n  }\n}\n")

func DefaultContext7ServerJSON() []byte {
	content := make([]byte, len(defaultContext7ServerJSON))
	copy(content, defaultContext7ServerJSON)
	return content
}

func DefaultContext7OverlayJSON() []byte {
	content := make([]byte, len(defaultContext7OverlayJSON))
	copy(content, defaultContext7OverlayJSON)
	return content
}

func OpenCodeContext7OverlayJSON() []byte {
	content := make([]byte, len(openCodeContext7OverlayJSON))
	copy(content, openCodeContext7OverlayJSON)
	return content
}

func OpenClawContext7OverlayJSON() []byte {
	content := make([]byte, len(openClawContext7OverlayJSON))
	copy(content, openClawContext7OverlayJSON)
	return content
}

func VSCodeContext7OverlayJSON() []byte {
	content := make([]byte, len(vsCodeContext7OverlayJSON))
	copy(content, vsCodeContext7OverlayJSON)
	return content
}

func AntigravityContext7OverlayJSON() []byte {
	content := make([]byte, len(antigravityContext7OverlayJSON))
	copy(content, antigravityContext7OverlayJSON)
	return content
}

func KimiContext7OverlayJSON() []byte {
	content := make([]byte, len(kimiContext7OverlayJSON))
	copy(content, kimiContext7OverlayJSON)
	return content
}
