package mcp

// Notion MCP server overlay variables.
// The Notion MCP server uses the official @notionhq/notion-mcp-server npm package.
// Auth tokens must be configured manually after installation — see post-install guidance.

var defaultNotionServerJSON = []byte(`{
  "command": "npx",
  "args": [
    "-y",
    "@notionhq/notion-mcp-server"
  ],
  "env": {
    "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
  }
}
`)

var defaultNotionOverlayJSON = []byte(`{
  "mcpServers": {
    "notion": {
      "command": "npx",
      "args": [
        "-y",
        "@notionhq/notion-mcp-server"
      ],
      "env": {
        "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
      }
    }
  }
}
`)

var openCodeNotionOverlayJSON = []byte(`{
  "mcp": {
    "notion": {
      "__replace__": {
        "type": "stdio",
        "command": "npx",
        "args": ["-y", "@notionhq/notion-mcp-server"],
        "env": {
          "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
        },
        "enabled": true
      }
    }
  }
}
`)

var openClawNotionOverlayJSON = []byte(`{
  "mcp": {
    "servers": {
      "notion": {
        "command": "npx",
        "args": [
          "-y",
          "@notionhq/notion-mcp-server"
        ],
        "env": {
          "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
        }
      }
    }
  }
}
`)

var vsCodeNotionOverlayJSON = []byte(`{
  "servers": {
    "notion": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@notionhq/notion-mcp-server"],
      "env": {
        "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
      }
    }
  }
}
`)

var antigravityNotionOverlayJSON = []byte(`{
  "mcpServers": {
    "notion": {
      "__replace__": {
        "command": "npx",
        "args": ["-y", "@notionhq/notion-mcp-server"],
        "env": {
          "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
        }
      }
    }
  }
}
`)

var kimiNotionOverlayJSON = []byte(`{
  "mcpServers": {
    "notion": {
      "__replace__": {
        "transport": "stdio",
        "command": "npx",
        "args": ["-y", "@notionhq/notion-mcp-server"],
        "env": {
          "OPENAPI_MCP_HEADERS": "Authorization: Bearer <your-notion-token>"
        }
      }
    }
  }
}
`)

// NotionAuthGuidance returns the post-install guidance message for Notion MCP auth.
// It includes the config path hint and documentation URL.
func NotionAuthGuidance(configPath string) string {
	return "Notion MCP server installed.\n" +
		"  Config file: " + configPath + "\n" +
		"  Next step: replace <your-notion-token> with your Notion integration token.\n" +
		"  Docs: https://developers.notion.com/docs/mcp"
}

func DefaultNotionServerJSON() []byte {
	content := make([]byte, len(defaultNotionServerJSON))
	copy(content, defaultNotionServerJSON)
	return content
}

func DefaultNotionOverlayJSON() []byte {
	content := make([]byte, len(defaultNotionOverlayJSON))
	copy(content, defaultNotionOverlayJSON)
	return content
}

func OpenCodeNotionOverlayJSON() []byte {
	content := make([]byte, len(openCodeNotionOverlayJSON))
	copy(content, openCodeNotionOverlayJSON)
	return content
}

func OpenClawNotionOverlayJSON() []byte {
	content := make([]byte, len(openClawNotionOverlayJSON))
	copy(content, openClawNotionOverlayJSON)
	return content
}

func VSCodeNotionOverlayJSON() []byte {
	content := make([]byte, len(vsCodeNotionOverlayJSON))
	copy(content, vsCodeNotionOverlayJSON)
	return content
}

func AntigravityNotionOverlayJSON() []byte {
	content := make([]byte, len(antigravityNotionOverlayJSON))
	copy(content, antigravityNotionOverlayJSON)
	return content
}

func KimiNotionOverlayJSON() []byte {
	content := make([]byte, len(kimiNotionOverlayJSON))
	copy(content, kimiNotionOverlayJSON)
	return content
}
